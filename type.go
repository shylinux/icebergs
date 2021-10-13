package ice

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	kit "shylinux.com/x/toolkits"
)

type Cache struct {
	Name  string
	Help  string
	Value string
}
type Config struct {
	Name  string
	Help  string
	Value interface{}
}
type Action struct {
	Name string
	Help string
	Hand func(m *Message, arg ...string)
	List []interface{}
}
type Command struct {
	Name   string
	Help   string
	Action map[string]*Action
	Meta   map[string]interface{}
	Hand   func(m *Message, c *Context, key string, arg ...string)
	List   []interface{}
}
type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
}
type Context struct {
	Name string
	Help string

	Caches   map[string]*Cache
	Configs  map[string]*Config
	Commands map[string]*Command

	contexts map[string]*Context
	context  *Context
	root     *Context

	begin  *Message
	start  *Message
	server Server

	wg *sync.WaitGroup
	id int32
}

func (c *Context) ID() int32 {
	if c == nil {
		return 1
	}
	return atomic.AddInt32(&c.id, 1)
}
func (c *Context) Cap(key string, arg ...interface{}) string {
	if len(arg) > 0 {
		c.Caches[key].Value = kit.Format(arg[0])
	}
	return c.Caches[key].Value
}
func (c *Context) Cmd(m *Message, key string, arg ...string) *Message {
	return c.cmd(m, m.target.Commands[key], key, arg...)
}
func (c *Context) Server() Server {
	return c.server
}

func (c *Context) Register(s *Context, x Server, name ...string) *Context {
	if s.Merge(s); s.Name == "" {
		s.Name = kit.Split(kit.Split(kit.FileLine(2, 3), ":")[0], "/")[1]
	}

	for _, n := range name {
		Name(n, s)
	}

	if c.contexts == nil {
		c.contexts = map[string]*Context{}
	}
	c.contexts[s.Name] = s
	s.root = c.root
	s.context = c
	s.server = x
	return s
}
func (c *Context) Merge(s *Context) *Context {
	if c.Commands == nil {
		c.Commands = map[string]*Command{}
	}
	for k, v := range s.Commands {
		if o, ok := c.Commands[k]; ok && s != c {
			func() {
				switch last, next := o.Hand, v.Hand; k {
				case CTX_INIT:
					v.Hand = func(m *Message, c *Context, key string, arg ...string) {
						last(m, c, key, arg...)
						next(m, c, key, arg...)
					}
				case CTX_EXIT:
					v.Hand = func(m *Message, c *Context, key string, arg ...string) {
						next(m, c, key, arg...)
						last(m, c, key, arg...)
					}
				}
			}()
		}

		if v.Meta == nil {
			v.Meta = kit.Dict()
		}
		if p := kit.Format(v.Meta[kit.MDB_STYLE]); p == "" {
			v.Meta[kit.MDB_STYLE] = k
		}
		if c.Commands[k] = v; v.List == nil {
			v.List = c.split(v.Name)
		}

		for k, a := range v.Action {
			help := strings.SplitN(a.Help, "：", 2)
			if len(help) == 1 || help[1] == "" {
				help = strings.SplitN(help[0], ":", 2)
			}
			kit.Value(v.Meta, kit.Keys("_trans", k), help[0])
			if len(help) > 1 {
				kit.Value(v.Meta, kit.Keys("title", k), help[1])
			}
			if a.Hand == nil {
				continue
			}
			if a.List == nil {
				a.List = c.split(a.Name)
			}
			if len(a.List) > 0 {
				v.Meta[k] = a.List
			}
		}
	}

	if c.Configs == nil {
		c.Configs = map[string]*Config{}
	}
	for k, v := range s.Configs {
		c.Configs[k] = v
	}

	if c.Caches == nil {
		c.Caches = map[string]*Cache{}
	}
	for k, v := range s.Caches {
		c.Caches[k] = v
	}
	return c
}

func (c *Context) Spawn(m *Message, name string, help string, arg ...string) *Context {
	s := &Context{Name: name, Help: help, Caches: map[string]*Cache{}, Configs: map[string]*Config{}}
	if m.target.server != nil {
		c.Register(s, m.target.server.Spawn(m, s, arg...))
	} else {
		c.Register(s, nil)
	}
	m.target = s
	return s
}
func (c *Context) Begin(m *Message, arg ...string) *Context {
	c.Caches[CTX_FOLLOW] = &Cache{Name: CTX_FOLLOW, Value: kit.Keys(kit.Select("", c.context.Cap(CTX_FOLLOW), c.context != Index), c.Name)}
	c.Caches[CTX_STATUS] = &Cache{Name: CTX_STATUS, Value: CTX_BEGIN}
	c.Caches[CTX_STREAM] = &Cache{Name: CTX_STREAM, Value: ""}
	m.Log(LOG_BEGIN, c.Cap(CTX_FOLLOW))

	if c.begin = m; c.server != nil {
		c.server.Begin(m, arg...)
	}
	return c
}
func (c *Context) Start(m *Message, arg ...string) bool {
	wait := make(chan bool)

	m.Hold(1)
	m.Go(func() {
		m.Log(LOG_START, c.Cap(CTX_FOLLOW))
		c.Cap(CTX_STATUS, CTX_START)
		wait <- true

		if c.start = m; c.server != nil {
			c.server.Start(m, arg...)
		}
		if m.Done(true); m.wait != nil {
			m.wait <- true
		}
	})

	<-wait
	return true
}
func (c *Context) Close(m *Message, arg ...string) bool {
	m.Log(LOG_CLOSE, c.Cap(CTX_FOLLOW))
	c.Cap(CTX_STATUS, CTX_CLOSE)

	if c.server != nil {
		return c.server.Close(m, arg...)
	}
	return true
}

type Message struct {
	time time.Time
	code int
	Hand bool
	wait chan bool

	meta map[string][]string
	data map[string]interface{}

	message *Message
	root    *Message

	source *Context
	target *Context
	_cmd   *Command
	_key   string

	cb func(*Message) *Message
	W  http.ResponseWriter
	R  *http.Request
	O  io.Writer
	I  io.Reader
}

func (m *Message) Time(args ...interface{}) string { // [duration] [format [args...]]
	t := m.time
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case string: // 时间偏移
			if d, e := time.ParseDuration(arg); e == nil {
				t, args = t.Add(d), args[1:]
			}
		}
	}
	f := MOD_TIME
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case string: // 时间格式
			if f = arg; len(args) > 1 {
				f = fmt.Sprintf(f, args[1:]...)
			}
		}
	}
	return t.Format(f)
}
func (m *Message) Target() *Context {
	return m.target
}
func (m *Message) Source() *Context {
	return m.source
}
func (m *Message) Spawn(arg ...interface{}) *Message {
	msg := &Message{
		time: time.Now(), code: int(m.target.root.ID()),
		meta: map[string][]string{},
		data: map[string]interface{}{},

		message: m, root: m.root,
		source: m.target, target: m.target,
		W: m.W, R: m.R, O: m.O, I: m.I,
	}

	if len(arg) > 0 {
		switch val := arg[0].(type) {
		case []byte:
			json.Unmarshal(val, &msg.meta)
		case *Context:
			msg.target = val
		}
	}
	return msg
}
func (m *Message) Start(key string, arg ...string) *Message {
	m.Search(key+".", func(p *Context, s *Context) {
		s.Start(m.Spawn(s), arg...)
	})
	return m
}
func (m *Message) Starts(name string, help string, arg ...string) *Message {
	m.wait = make(chan bool)
	m.target.Spawn(m, name, help, arg...).Begin(m, arg...).Start(m, arg...)
	<-m.wait
	return m
}
func (m *Message) Travel(cb interface{}) *Message {
	list := []*Context{m.root.target}
	for i := 0; i < len(list); i++ {
		switch cb := cb.(type) {
		case func(*Context, *Context):
			// 遍历模块
			cb(list[i].context, list[i])
		case func(*Context, *Context, string, *Command):
			ls := []string{}
			for k := range list[i].Commands {
				ls = append(ls, k)
			}
			sort.Strings(ls)

			target := m.target
			for _, k := range ls { // 命令列表
				m.target = list[i]
				cb(list[i].context, list[i], k, list[i].Commands[k])
			}
			m.target = target
		case func(*Context, *Context, string, *Config):
			ls := []string{}
			for k := range list[i].Configs {
				ls = append(ls, k)
			}
			sort.Strings(ls)

			for _, k := range ls { // 配置列表
				cb(list[i].context, list[i], k, list[i].Configs[k])
			}
		}

		ls := []string{}
		for k := range list[i].contexts {
			ls = append(ls, k)
		}
		sort.Strings(ls)

		for _, k := range ls { // 遍历递进
			list = append(list, list[i].contexts[k])
		}
	}
	return m
}
func (m *Message) Search(key string, cb interface{}) *Message {
	if key == "" {
		return m
	}

	// 查找模块
	p := m.target.root
	if ctx, ok := Info.names[key].(*Context); ok {
		p = ctx
	} else if key == "ice." {
		p, key = m.target.root, ""
	} else if key == "." {
		p, key = m.target, ""
	} else if key == ".." {
		if m.target.context != nil {
			p, key = m.target.context, ""
		}
	} else if strings.Contains(key, ".") {
		list := strings.Split(key, ".")
		for _, p = range []*Context{m.target.root, m.target, m.source} {
			if p == nil {
				continue
			}
			for _, v := range list[:len(list)-1] {
				if s, ok := p.contexts[v]; ok {
					p = s
				} else {
					p = nil
					break
				}
			}
			if p != nil {
				break
			}
		}
		if m.Warn(p == nil, ErrNotFound, key) {
			return m
		}
		key = list[len(list)-1]
	} else {
		p = m.target
	}

	// 遍历命令
	switch cb := cb.(type) {
	case func(key string, cmd *Command):
		if key == "" {
			for k, v := range p.Commands {
				cb(k, v)
			}
		} else if cmd, ok := p.Commands[key]; ok {
			cb(key, cmd)
		}

	case func(p *Context, s *Context, key string, cmd *Command):
		if key == "" {
			for k, v := range p.Commands {
				cb(p.context, p, k, v)
			}
			break
		}

		// 查找命令
		for _, p := range []*Context{p, m.target, m.source} {
			for s := p; s != nil; s = s.context {
				if cmd, ok := s.Commands[key]; ok {
					cb(s.context, s, key, cmd)
					return m
				}
			}
		}
	case func(p *Context, s *Context, key string, conf *Config):
		// 查找配置
		for _, p := range []*Context{m.target, p, m.source} {
			for s := p; s != nil; s = s.context {
				if cmd, ok := s.Configs[key]; ok {
					cb(s.context, s, key, cmd)
					return m
				}
			}
		}
	case func(p *Context, s *Context, key string):
		cb(p.context, p, key)
	case func(p *Context, s *Context):
		cb(p.context, p)
	}
	return m
}

func (m *Message) Cmdy(arg ...interface{}) *Message {
	return m.Copy(m.cmd(arg...))
}
func (m *Message) Cmdx(arg ...interface{}) string {
	return kit.Select("", m.cmd(arg...).meta[MSG_RESULT], 0)
}
func (m *Message) Cmds(arg ...interface{}) *Message {
	return m.Go(func() { m.cmd(arg...) })
}
func (m *Message) Cmd(arg ...interface{}) *Message {
	return m.cmd(arg...)
}

func (m *Message) Confi(key string, sub string) int {
	return kit.Int(m.Conf(key, sub))
}
func (m *Message) Confv(arg ...interface{}) (val interface{}) {
	m.Search(kit.Format(arg[0]), func(p *Context, s *Context, key string, conf *Config) {
		if len(arg) == 1 {
			val = conf.Value
			return // 读配置
		}

		if len(arg) > 2 {
			if arg[1] == nil || arg[1] == "" {
				conf.Value = arg[2] // 写配置
			} else {
				kit.Value(conf.Value, arg[1:]...) // 写配置项
			}
		}
		val = kit.Value(conf.Value, arg[1]) // 读配置项
	})
	return
}
func (m *Message) Confm(key string, chain interface{}, cbs ...interface{}) map[string]interface{} {
	val := m.Confv(key, chain)
	if len(cbs) > 0 {
		kit.Fetch(val, cbs[0])
	}
	value, _ := val.(map[string]interface{})
	return value
}
func (m *Message) Conf(arg ...interface{}) string {
	return kit.Format(m.Confv(arg...))
}
func (m *Message) Capi(key string, val ...interface{}) int {
	if len(val) > 0 {
		m.Cap(key, kit.Int(m.Cap(key))+kit.Int(val[0]))
	}
	return kit.Int(m.Cap(key))
}
func (m *Message) Capv(arg ...interface{}) interface{} {
	key := ""
	switch val := arg[0].(type) {
	case string:
		key, arg = val, arg[1:]
	}

	for _, s := range []*Context{m.target} {
		for c := s; c != nil; c = c.context {
			if caps, ok := c.Caches[key]; ok {
				if len(arg) > 0 {
					// 写数据
					caps.Value = kit.Format(arg[0])
				}
				// 读数据
				return caps.Value
			}
		}
	}
	return nil
}
func (m *Message) Cap(arg ...interface{}) string {
	return kit.Format(m.Capv(arg...))
}
