package ice

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	kit "github.com/shylinux/toolkits"
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
	List []interface{}
	Hand func(m *Message, arg ...string)
}
type Command struct {
	Name   string
	Help   string
	List   []interface{}
	Meta   map[string]interface{}
	Hand   func(m *Message, c *Context, key string, arg ...string)
	Action map[string]*Action
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
type Server interface {
	Spawn(m *Message, c *Context, arg ...string) Server
	Begin(m *Message, arg ...string) Server
	Start(m *Message, arg ...string) bool
	Close(m *Message, arg ...string) bool
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
func (c *Context) _cmd(m *Message, cmd *Command, key string, k string, h *Action, arg ...string) *Message {
	if k == "run" && m.Warn(!m.Right(arg), ErrNotRight, arg) {
		return m
	}

	m.Log(LOG_CMDS, "%s.%s %s %d %v %s", c.Name, key, k, len(arg), arg, kit.FileLine(h.Hand, 3))
	if len(h.List) > 0 && k != "search" {
		order := false
		for i, v := range h.List {
			name := kit.Format(kit.Value(v, "name"))
			value := kit.Format(kit.Value(v, "value"))

			if i == 0 && len(arg) > 0 && arg[0] != name {
				order = true
			}
			if order {
				value = kit.Select(value, arg, i)
			}

			m.Option(name, kit.Select(m.Option(name), value, !strings.HasPrefix(value, "@")))
		}
		if !order {
			for i := 0; i < len(arg)-1; i += 2 {
				m.Option(arg[i], arg[i+1])
			}
		}
	}

	if h.Hand == nil {
		m.Cmdy(kit.Split(h.Name), arg)
	} else {
		h.Hand(m, arg...)
	}
	return m
}
func (c *Context) cmd(m *Message, cmd *Command, key string, arg ...string) *Message {
	if m._key, m._cmd = key, cmd; cmd == nil {
		return m
	}

	m.meta[MSG_DETAIL] = kit.Simple(key, arg)
	if m.Hand = true; len(arg) > 1 && arg[0] == kit.MDB_ACTION && cmd.Action != nil {
		if h, ok := cmd.Action[arg[1]]; ok {
			return c._cmd(m, cmd, key, arg[1], h, arg[2:]...)
		}
	}
	if len(arg) > 0 && arg[0] != "command" && cmd.Action != nil {
		if h, ok := cmd.Action[arg[0]]; ok {
			return c._cmd(m, cmd, key, arg[0], h, arg[1:]...)
		}
	}

	m.Log(LOG_CMDS, "%s.%s %d %v %s", c.Name, key, len(arg), arg,
		kit.Select(kit.FileLine(cmd.Hand, 3), kit.FileLine(9, 3), m.target.Name == "mdb"))
	cmd.Hand(m, c, key, arg...)
	return m
}
func (c *Context) Cmd(m *Message, cmd string, key string, arg ...string) *Message {
	return c.cmd(m, m.Target().Commands[cmd], cmd, arg...)
}
func (c *Context) Server() Server {
	return c.server
}

func (c *Context) Register(s *Context, x Server, name ...string) *Context {
	if s.Name == "" {
		s.Name = kit.Split(kit.Split(kit.FileLine(2, 3), ":")[0], "/")[1]
	}
	for _, n := range name {
		Name(n, s)
	}
	s.Merge(s)

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

		c.Commands[k] = v

		if v.List == nil {
			v.List = c.split(k, v, v.Name)
		}
		if v.Meta == nil {
			v.Meta = kit.Dict()
		}

		if p := kit.Format(v.Meta[kit.MDB_STYLE]); p == "" {
			v.Meta[kit.MDB_STYLE] = k
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
				a.List = c.split(k, nil, a.Name)
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
func (c *Context) split(key string, cmd *Command, name string) []interface{} {
	const (
		BUTTON   = "button"
		SELECT   = "select"
		TEXT     = "text"
		TEXTAREA = "textarea"
	)

	button, list := false, []interface{}{}
	for _, v := range kit.Split(kit.Select("key", name), " ", " ")[1:] {
		switch v {
		case "text":
			list = append(list, kit.List(kit.MDB_INPUT, TEXTAREA, kit.MDB_NAME, "text")...)
			continue
		case "page":
			list = append(list, kit.List(kit.MDB_INPUT, TEXT, kit.MDB_NAME, "limit")...)
			list = append(list, kit.List(kit.MDB_INPUT, TEXT, kit.MDB_NAME, "offend")...)
			list = append(list, kit.List(kit.MDB_INPUT, BUTTON, kit.MDB_NAME, "prev")...)
			list = append(list, kit.List(kit.MDB_INPUT, BUTTON, kit.MDB_NAME, "next")...)
			continue
		case "auto":
			list = append(list, kit.List(kit.MDB_INPUT, BUTTON, kit.MDB_NAME, "查看", kit.MDB_VALUE, "auto")...)
			list = append(list, kit.List(kit.MDB_INPUT, BUTTON, kit.MDB_NAME, "返回")...)
			button = true
			continue
		}

		ls, value := kit.Split(v, " ", ":=@"), ""
		item := kit.Dict(kit.MDB_INPUT, kit.Select(TEXT, BUTTON, button))
		if kit.Value(item, kit.MDB_NAME, ls[0]); item[kit.MDB_INPUT] == TEXT {
			kit.Value(item, kit.MDB_VALUE, kit.Select("@key", "auto", strings.Contains(name, "auto")))
		}

		for i := 1; i < len(ls); i += 2 {
			switch ls[i] {
			case ":":
				switch kit.Value(item, kit.MDB_INPUT, ls[i+1]); ls[i+1] {
				case TEXTAREA:
					kit.Value(item, "style.width", "360")
					kit.Value(item, "style.height", "60")
				case BUTTON:
					kit.Value(item, kit.MDB_VALUE, "")
					button = true
				}
			case "=":
				if value = kit.Select("", ls, i+1); len(ls) > i+1 && strings.Contains(ls[i+1], ",") {
					vs := strings.Split(ls[i+1], ",")
					kit.Value(item, "values", vs)
					kit.Value(item, kit.MDB_VALUE, vs[0])
					kit.Value(item, kit.MDB_INPUT, SELECT)
					if kit.Value(item, kit.MDB_NAME) == "scale" {
						kit.Value(item, kit.MDB_VALUE, "week")
					}
				} else {
					kit.Value(item, kit.MDB_VALUE, value)
				}
			case "@":
				if len(ls) > i+1 {
					if kit.Value(item, kit.MDB_INPUT) == BUTTON {
						kit.Value(item, kit.MDB_ACTION, ls[i+1])
					} else {
						kit.Value(item, kit.MDB_VALUE, "@"+ls[i+1]+"="+value)
					}
				}
			}
		}
		list = append(list, item)
	}
	return list
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
		case string:
			if d, e := time.ParseDuration(arg); e == nil {
				// 时间偏移
				t, args = t.Add(d), args[1:]
			}
		}
	}
	f := MOD_TIME
	if len(args) > 0 {
		switch arg := args[0].(type) {
		case string:
			if f = arg; len(args) > 1 {
				// 时间格式
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
func (m *Message) Format(key interface{}) string {
	switch key := key.(type) {
	case []byte:
		json.Unmarshal(key, &m.meta)
	case string:
		switch key {
		case "cost":
			return kit.FmtTime(kit.Int64(time.Since(m.time)))
		case "meta":
			return kit.Format(m.meta)
		case "size":
			if len(m.meta["append"]) == 0 {
				return fmt.Sprintf("%dx%d", 0, 0)
			} else {
				return fmt.Sprintf("%dx%d", len(m.meta[m.meta["append"][0]]), len(m.meta["append"]))
			}
		case "append":
			if len(m.meta["append"]) == 0 {
				return fmt.Sprintf("%dx%d %s", 0, 0, "[]")
			} else {
				return fmt.Sprintf("%dx%d %s", len(m.meta[m.meta["append"][0]]), len(m.meta["append"]), kit.Format(m.meta["append"]))
			}

		case "time":
			return m.Time()
		case "ship":
			return fmt.Sprintf("%s->%s", m.source.Name, m.target.Name)
		case "prefix":
			return fmt.Sprintf("%s %d %s->%s", m.Time(), m.code, m.source.Name, m.target.Name)

		case "chain":
			// 调用链
			ms := []*Message{}
			for msg := m; msg != nil; msg = msg.message {
				ms = append(ms, msg)
			}

			meta := append([]string{}, "\n\n")
			for i := len(ms) - 1; i >= 0; i-- {
				msg := ms[i]

				meta = append(meta, fmt.Sprintf("%s ", msg.Format("prefix")))
				if len(msg.meta[MSG_DETAIL]) > 0 {
					meta = append(meta, fmt.Sprintf("detail:%d %v", len(msg.meta[MSG_DETAIL]), msg.meta[MSG_DETAIL]))
				}

				if len(msg.meta[MSG_OPTION]) > 0 {
					meta = append(meta, fmt.Sprintf("option:%d %v\n", len(msg.meta[MSG_OPTION]), msg.meta[MSG_OPTION]))
					for _, k := range msg.meta[MSG_OPTION] {
						if v, ok := msg.meta[k]; ok {
							meta = append(meta, fmt.Sprintf("    %s: %d %v\n", k, len(v), v))
						}
					}
				} else {
					meta = append(meta, "\n")
				}

				if len(msg.meta[MSG_APPEND]) > 0 {
					meta = append(meta, fmt.Sprintf("  append:%d %v\n", len(msg.meta[MSG_APPEND]), msg.meta[MSG_APPEND]))
					for _, k := range msg.meta[MSG_APPEND] {
						if v, ok := msg.meta[k]; ok {
							meta = append(meta, fmt.Sprintf("    %s: %d %v\n", k, len(v), v))
						}
					}
				}
				if len(msg.meta[MSG_RESULT]) > 0 {
					meta = append(meta, fmt.Sprintf("  result:%d %v\n", len(msg.meta[MSG_RESULT]), msg.meta[MSG_RESULT]))
				}
			}
			return strings.Join(meta, "")
		case "stack":
			// 调用栈
			pc := make([]uintptr, 100)
			pc = pc[:runtime.Callers(5, pc)]
			frames := runtime.CallersFrames(pc)

			meta := []string{}
			for {
				frame, more := frames.Next()
				file := strings.Split(frame.File, "/")
				name := strings.Split(frame.Function, "/")
				meta = append(meta, fmt.Sprintf("\n%s:%d\t%s", file[len(file)-1], frame.Line, name[len(name)-1]))
				if !more {
					break
				}
			}
			return strings.Join(meta, "")
		}
	}
	return m.time.Format(MOD_TIME)
}
func (m *Message) Formats(key string) string {
	switch key {
	case "meta":
		return kit.Formats(m.meta)
	}
	return m.Format(key)
}
func (m *Message) Spawn(arg ...interface{}) *Message {
	msg := &Message{
		time: time.Now(),
		code: int(m.target.root.ID()),

		meta: map[string][]string{},
		data: map[string]interface{}{},

		message: m, root: m.root,

		source: m.target,
		target: m.target,

		W: m.W, R: m.R,
		O: m.O, I: m.I,
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
	list := []*Context{m.target}
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
func (m *Message) cmd(arg ...interface{}) *Message {
	opts := map[string]interface{}{}
	args := []interface{}{}
	var cbs interface{}

	// 解析参数
	for _, v := range arg {
		switch val := v.(type) {
		case func(int, map[string]string, []string):
			defer func() { m.Table(val) }()

		case map[string]string:
			for k, v := range val {
				opts[k] = v
			}

		case *Option:
			opts[val.Name] = val.Value
		case Option:
			opts[val.Name] = val.Value

		case *Sort:
			defer func() { m.Sort(val.Fields, val.Method) }()
		case Sort:
			defer func() { m.Sort(val.Fields, val.Method) }()

		default:
			if reflect.Func == reflect.TypeOf(val).Kind() {
				cbs = val
			} else {
				args = append(args, v)
			}
		}
	}

	// 解析命令
	list := kit.Simple(args...)
	if len(list) == 0 && !m.Hand {
		list = m.meta[MSG_DETAIL]
	}
	if len(list) == 0 {
		return m
	}

	ok := false
	run := func(msg *Message, ctx *Context, cmd *Command, key string, arg ...string) {
		if ok = true; cbs != nil {
			msg.Option(list[0]+".cb", cbs)
		}
		for k, v := range opts {
			msg.Option(k, v)
		}

		// 执行命令
		m.TryCatch(msg, true, func(msg *Message) {
			m = ctx.cmd(msg, cmd, key, arg...)
		})
	}

	// 查找命令
	if cmd, ok := m.target.Commands[list[0]]; ok {
		run(m.Spawn(), m.target, cmd, list[0], list[1:]...)
	} else if cmd, ok := m.source.Commands[list[0]]; ok {
		run(m.Spawn(m.source), m.source, cmd, list[0], list[1:]...)
	} else {
		m.Search(list[0], func(p *Context, s *Context, key string, cmd *Command) {
			run(m.Spawn(s), s, cmd, key, list[1:]...)
		})
	}

	// 系统命令
	if m.Warn(!ok, ErrNotFound, list) {
		return m.Set(MSG_RESULT).Cmdy("cli.system", list)
	}
	return m
}

func (m *Message) Confv(arg ...interface{}) (val interface{}) {
	m.Search(kit.Format(arg[0]), func(p *Context, s *Context, key string, conf *Config) {
		if len(arg) == 1 {
			val = conf.Value
			return // 读配置
		}

		if len(arg) > 2 {
			if arg[1] == nil || arg[1] == "" {
				// 写配置
				conf.Value = arg[2]
			} else {
				// 写修改项
				kit.Value(conf.Value, arg[1:]...)
			}
		}
		// 读配置项
		val = kit.Value(conf.Value, arg[1])
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
