package ice

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync/atomic"
	"time"

	kit "shylinux.com/x/toolkits"
)

type Any = interface{}
type List = []Any
type Map = map[string]Any
type Maps = map[string]string
type Handler func(m *Message, arg ...string)
type Commands = map[string]*Command
type Actions = map[string]*Action
type Configs = map[string]*Config
type Caches = map[string]*Cache
type Contexts = map[string]*Context
type Messages = map[string]*Message

type Cache struct {
	Name  string
	Help  string
	Value string
}
type Config struct {
	Name  string
	Help  string
	Value Any
}
type Action struct {
	Name string
	Help string
	Hand Handler
	List List
}
type Command struct {
	Name    string
	Help    string
	Actions Actions
	Hand    Handler
	RawHand Any
	List    List
	Meta    Map
}
type Context struct {
	Name string
	Help string

	Caches   Caches
	Configs  Configs
	Commands Commands

	Contexts Contexts
	context  *Context
	root     *Context
	server   Server

	begin *Message
	start *Message

	id int32
}
type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
}

func (c *Context) ID() int32 {
	return atomic.AddInt32(&c.id, 1)
}
func (c *Command) GetFileLine() string {
	if c.RawHand != nil {
		switch h := c.RawHand.(type) {
		case string:
			return h
		default:
			return kit.FileLine(c.RawHand, 100)
		}
	} else if c.Hand != nil {
		return kit.FileLine(c.Hand, 100)
	} else {
		return ""
	}
}
func (c *Context) Cap(key string, arg ...Any) string {
	if len(arg) > 0 {
		c.Caches[key].Value = kit.Format(arg[0])
	}
	return c.Caches[key].Value
}
func (c *Context) Cmd(m *Message, key string, arg ...string) *Message {
	return c._command(m, c.Commands[key], key, arg...)
}
func (c *Context) Server() Server {
	return c.server
}

func (c *Context) PrefixKey(arg ...string) string {
	return kit.Keys(c.Cap(CTX_FOLLOW), arg)
}
func (c *Context) Register(s *Context, x Server, n ...string) *Context {
	for _, n := range n {
		if s, ok := Info.names[n]; ok {
			last := ""
			switch s := s.(type) {
			case *Context:
				last = s.Name
			}
			panic(kit.Format("%s %s %v", ErrWarn, n, last))
		}
		Info.names[n] = s
	}

	if s.Merge(s); c.Contexts == nil {
		c.Contexts = Contexts{}
	}
	c.Contexts[s.Name] = s
	s.root = c.root
	s.context = c
	s.server = x
	return s
}
func (c *Context) MergeCommands(Commands Commands) *Context {
	configs := Configs{}
	for k, _ := range Commands {
		configs[k] = &Config{Value: kit.Data()}
	}
	return c.Merge(&Context{Commands: Commands, Configs: configs})
}
func (c *Context) Merge(s *Context) *Context {
	if c.Commands == nil {
		c.Commands = Commands{}
	}
	if c.Commands[CTX_INIT] == nil {
		c.Commands[CTX_INIT] = &Command{Hand: func(m *Message, arg ...string) { Info.Load(m) }}
	}
	if c.Commands[CTX_EXIT] == nil {
		c.Commands[CTX_EXIT] = &Command{Hand: func(m *Message, arg ...string) { Info.Save(m) }}
	}

	merge := func(pre *Command, before bool, key string, cmd *Command, cb ...Handler) {
		last := pre.Hand
		pre.Hand = func(m *Message, arg ...string) {
			if before {
				last(m, arg...)
			}

			_key, _cmd := m._key, m._cmd
			m._key, m._cmd = key, cmd
			for _, cb := range cb {
				if cb != nil {
					cb(m, arg...)
				}
			}
			m._key, m._cmd = _key, _cmd
			if !before {
				last(m, arg...)
			}
		}
	}

	for key, cmd := range s.Commands {
		if p, ok := c.Commands[key]; ok && s != c {
			switch hand := cmd.Hand; key {
			case CTX_INIT:
				merge(p, true, key, cmd, hand)
			case CTX_EXIT:
				merge(p, false, key, cmd, hand)
			}
		}

		if c.Commands[key] = cmd; cmd.List == nil {
			cmd.List = SplitCmd(cmd.Name, cmd.Actions)
		}
		if cmd.Meta == nil {
			cmd.Meta = kit.Dict()
		}

		for k, a := range cmd.Actions {
			if p, ok := c.Commands[k]; ok {
				switch h := a.Hand; k {
				case CTX_INIT:
					merge(p, true, key, cmd, func(m *Message, arg ...string) { h(m, arg...) })
				case CTX_EXIT:
					merge(p, false, key, cmd, func(m *Message, arg ...string) { h(m, arg...) })
				}
			}
			if strings.HasPrefix(k, PS) {
				k = kit.Select(k, path.Join(PS, key)+PS, k == PS)
				c.Commands[k] = &Command{Name: k, Help: cmd.Help, Hand: func(m *Message, arg ...string) { m.Cmdy(m.CommandKey(), arg) }}
				continue
			}

			if s != c {
				switch k {
				case "search":
					merge(c.Commands[CTX_INIT], true, key, cmd, func(m *Message, arg ...string) {
						if m.CommandKey() != "search" {
							m.Cmd("search", "create", m.CommandKey(), m.PrefixKey())
						}
					})
				}
			}

			help := strings.SplitN(a.Help, "：", 2)
			if len(help) == 1 || help[1] == "" {
				help = strings.SplitN(help[0], ":", 2)
			}
			if kit.Value(cmd.Meta, kit.Keys("_trans", strings.TrimPrefix(k, "_")), help[0]); len(help) > 1 {
				kit.Value(cmd.Meta, kit.Keys("title", k), help[1])
			}
			if a.Hand == nil {
				continue // alias cmd
			}
			if a.List == nil {
				a.List = SplitCmd(a.Name, nil)
			}
			if len(a.List) > 0 {
				cmd.Meta[k] = a.List
			}
		}
		delete(cmd.Actions, CTX_INIT)
		delete(cmd.Actions, CTX_EXIT)
	}

	if c.Configs == nil {
		c.Configs = Configs{}
	}
	for k, v := range s.Configs {
		c.Configs[k] = v
	}
	if c.Caches == nil {
		c.Caches = Caches{}
	}
	for k, v := range s.Caches {
		c.Caches[k] = v
	}
	return c
}
func (c *Context) Spawn(m *Message, name string, help string, arg ...string) *Context {
	s := &Context{Name: name, Help: help}
	if c.server != nil {
		c.Register(s, c.server.Spawn(m, s, arg...))
	} else {
		c.Register(s, nil)
	}
	m.target = s
	return s
}
func (c *Context) Begin(m *Message, arg ...string) *Context {
	follow := c.Name
	if c.context != nil && c.context != Index {
		follow = kit.Keys(c.context.Cap(CTX_FOLLOW), c.Name)
	}
	c.Caches[CTX_FOLLOW] = &Cache{Name: CTX_FOLLOW, Value: follow}
	c.Caches[CTX_STATUS] = &Cache{Name: CTX_STATUS, Value: CTX_BEGIN}
	c.Caches[CTX_STREAM] = &Cache{Name: CTX_STREAM, Value: ""}

	if c.begin = m; c.server != nil {
		c.server.Begin(m, arg...)
	}
	return c
}
func (c *Context) Start(m *Message, arg ...string) bool {
	wait := make(chan bool, 1)
	defer func() { <-wait }()

	m.Go(func() {
		m.Log(CTX_START, c.Cap(CTX_FOLLOW))
		c.Cap(CTX_STATUS, CTX_START)
		wait <- true

		if c.start = m; c.server != nil {
			c.server.Start(m, arg...)
		}
	})
	return true
}
func (c *Context) Close(m *Message, arg ...string) bool {
	m.Log(CTX_CLOSE, c.Cap(CTX_FOLLOW))
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

	meta map[string][]string
	data Map

	message *Message
	root    *Message

	_source string
	_target string
	source  *Context
	target  *Context
	_cmd    *Command
	_key    string
	_sub    string

	W http.ResponseWriter
	R *http.Request
	O io.Writer
	I io.Reader
}

// func (m *Message) getMeta(key string) []string {
// 	defer m.lock.RLock()()
// 	return m.meta[key]
// }
// func (m *Message) setMeta(key string, value []string) {
// 	defer m.lock.Lock()()
// 	if value == nil {
// 		delete(m.meta, key)
// 	} else {
// 		m.meta[key] = value
// 	}
// }
// func (m *Message) getData(key string) Any {
// 	defer m.lock.RLock()()
// 	return m._data[key]
// }
//
func (m *Message) Time(args ...Any) string { // [duration] [format [args...]]
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
func (m *Message) Spawn(arg ...Any) *Message {
	msg := &Message{
		time: time.Now(), code: int(m.target.root.ID()),
		meta: map[string][]string{}, data: Map{},
		message: m, root: m.root,
		source: m.target, target: m.target, _cmd: m._cmd, _key: m._key, _sub: m._sub,
		W: m.W, R: m.R, O: m.O, I: m.I,
	}

	for _, val := range arg {
		switch val := val.(type) {
		case []byte:
			json.Unmarshal(val, &msg.meta)
		case Option:
			msg.Option(val.Name, val.Value)
		case Map:
			for k, v := range val {
				msg.Option(k, v)
			}
		case Maps:
			for k, v := range val {
				msg.Option(k, v)
			}
		case http.ResponseWriter:
			msg.W = val
		case *http.Request:
			msg.R = val
		case *Context:
			msg.target = val
		case string:
			msg._key = val
		}
	}
	return msg
}
func (m *Message) Start(key string, arg ...string) *Message {
	m.Search(key+PT, func(p *Context, s *Context) { s.Start(m.Spawn(s), arg...) })
	return m
}
func (m *Message) Travel(cb Any) *Message {
	list := []*Context{m.root.target}
	for i := 0; i < len(list); i++ {
		switch cb := cb.(type) {
		case func(*Context, *Context): // 遍历模块
			cb(list[i].context, list[i])

		case func(*Context, *Context, string, *Command):
			target := m.target
			for _, k := range kit.SortedKey(list[i].Commands) { // 命令列表
				m.target = list[i]
				cb(list[i].context, list[i], k, list[i].Commands[k])
			}
			m.target = target

		case func(*Context, *Context, string, *Config):
			target := m.target
			for _, k := range kit.SortedKey(list[i].Configs) { // 配置列表
				m.target = list[i]
				cb(list[i].context, list[i], k, list[i].Configs[k])
			}
			m.target = target
		default:
			m.ErrorNotImplement(cb)
		}

		for _, k := range kit.SortedKey(list[i].Contexts) { // 遍历递进
			list = append(list, list[i].Contexts[k])
		}
	}
	return m
}
func (m *Message) Search(key string, cb Any) *Message {
	if key == "" {
		return m
	}

	// 查找模块
	p := m.target.root
	if key = strings.TrimPrefix(key, "ice."); key == PT {
		p, key = m.target, ""
	} else if key == ".." {
		p, key = m.target.context, ""
	} else if key == "ice." {
		p, key = m.target.root, ""
	} else if strings.Contains(key, PT) {
		ls := strings.Split(key, PT)
		for _, p = range []*Context{m.target.root, m.target, m.source} {
			if p == nil {
				continue
			}
			for _, k := range ls[:len(ls)-1] {
				if p = p.Contexts[k]; p == nil {
					break
				}
			}
			if p != nil {
				break
			}
		}
		if p == nil {
			return m
		}
		key = ls[len(ls)-1]
	} else if ctx, ok := Info.names[key].(*Context); ok {
		p = ctx
	} else {
		p = m.target
	}

	switch cb := cb.(type) {
	case func(key string, cmd *Command):
		if key == "" {
			for k, v := range p.Commands {
				cb(k, v) // 遍历命令
			}
			break
		}

		if cmd, ok := p.Commands[key]; ok {
			cb(key, cmd) // 查找命令
		}

	case func(p *Context, s *Context, key string, cmd *Command):
		if key == "" {
			for k, v := range p.Commands {
				cb(p.context, p, k, v) // 遍历命令
			}
			break
		}

		for _, p := range []*Context{p, m.target, m.source} {
			for s := p; s != nil; s = s.context {
				if cmd, ok := s.Commands[key]; ok {
					cb(s.context, s, key, cmd) // 查找命令
					return m
				}
			}
		}
	case func(p *Context, s *Context, key string, conf *Config):
		if key == "" {
			for k, v := range p.Configs {
				cb(p.context, p, k, v) // 遍历命令
			}
			break
		}

		for _, p := range []*Context{p, m.target, m.source} {
			for s := p; s != nil; s = s.context {
				if cmd, ok := s.Configs[key]; ok {
					cb(s.context, s, key, cmd) // 查找配置
					return m
				}
			}
		}
	case func(p *Context, s *Context, key string):
		cb(p.context, p, key) // 查找模块
	case func(p *Context, s *Context):
		cb(p.context, p) // 查找模块
	default:
		m.ErrorNotImplement(cb)
	}
	return m
}

func (m *Message) Commands(key string) *Command {
	return m.Target().Commands[key]
}
func (m *Message) Actions(key string) *Action {
	return m._cmd.Actions[key]
}
func (m *Message) CmdAppend(arg ...string) string {
	field := kit.Slice(arg, -1)[0]
	return m._command(kit.Slice(arg, 0, -1), OptionFields(field)).Append(field)
}
func (m *Message) CmdMap(arg ...string) map[string]map[string]string {
	field, list := kit.Slice(arg, -1)[0], map[string]map[string]string{}
	m._command(kit.Slice(arg, 0, -1)).Tables(func(value Maps) {
		list[value[field]] = value
	})
	return list
}
func (m *Message) Cmd(arg ...Any) *Message {
	return m._command(arg...)
}
func (m *Message) Cmds(arg ...Any) *Message {
	return m.Go(func() { m._command(arg...) })
}
func (m *Message) Cmdx(arg ...Any) string {
	res := kit.Select("", m._command(arg...).meta[MSG_RESULT], 0)
	return kit.Select("", res, res != ErrWarn)
}
func (m *Message) Cmdy(arg ...Any) *Message {
	return m.Copy(m._command(arg...))
}
func (m *Message) Confi(key string, sub string) int {
	return kit.Int(m.Conf(key, sub))
}
func (m *Message) Confv(arg ...Any) (val Any) { // key sub val
	run := func(conf *Config) {
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
	}

	key := kit.Format(arg[0])
	if key == "" {
		key = m._key
	}
	if conf, ok := m.target.Configs[key]; ok {
		run(conf)
	} else if conf, ok := m.source.Configs[key]; ok {
		run(conf)
	} else {
		m.Search(key, func(p *Context, s *Context, key string, conf *Config) { run(conf) })
	}
	return
}
func (m *Message) Confm(key string, sub Any, cbs ...Any) Map {
	val := m.Confv(key, sub)
	if len(cbs) > 0 {
		kit.Fetch(val, cbs[0])
	}
	value, _ := val.(Map)
	return value
}
func (m *Message) Conf(arg ...Any) string { // key sub val
	return kit.Format(m.Confv(arg...))
}
func (m *Message) Capi(key string, val ...Any) int {
	if len(val) > 0 {
		m.Cap(key, kit.Int(m.Cap(key))+kit.Int(val[0]))
	}
	return kit.Int(m.Cap(key))
}
func (m *Message) Capv(arg ...Any) Any {
	key := ""
	switch val := arg[0].(type) {
	case string:
		key, arg = val, arg[1:]
	}

	for _, s := range []*Context{m.target} {
		for c := s; c != nil; c = c.context {
			if caps, ok := c.Caches[key]; ok {
				if len(arg) > 0 { // 写数据
					caps.Value = kit.Format(arg[0])
				}
				return caps.Value // 读数据
			}
		}
	}
	return nil
}
func (m *Message) Cap(arg ...Any) string {
	return kit.Format(m.Capv(arg...))
}
