package ice

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
	"shylinux.com/x/toolkits/task"
)

type Any = interface{}
type List = []Any
type Map = map[string]Any
type Maps = map[string]string
type Handler func(m *Message, arg ...string)
type Messages = map[string]*Message
type Contexts = map[string]*Context
type Commands = map[string]*Command
type Actions = map[string]*Action
type Configs = map[string]*Config
type Caches = map[string]*Cache

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
	Icon string
	Hand Handler
	List List
}
type Command struct {
	Name    string
	Help    string
	Icon    string
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

	contexts Contexts
	context  *Context
	root     *Context
	server   Server

	id int32
}
type Server interface {
	Begin(m *Message, arg ...string)
	Start(m *Message, arg ...string)
	Close(m *Message, arg ...string)
}

func (c *Context) Cap(key string, arg ...Any) string {
	kit.If(len(arg) > 0, func() { c.Caches[key].Value = kit.Format(arg[0]) })
	return c.Caches[key].Value
}
func (c *Context) Cmd(m *Message, key string, arg ...string) *Message {
	return c._command(m, c.Commands[key], key, arg...)
}
func (c *Context) Prefix(arg ...string) string {
	return kit.Keys(c.Cap(CTX_FOLLOW), arg)
}
func (c *Context) Server() Server { return c.server }
func (c *Context) ID() int32      { return atomic.AddInt32(&c.id, 1) }
func (c *Context) Register(s *Context, x Server, cmd ...string) *Context {
	kit.For(cmd, func(cmd string) {
		if s, ok := Info.Index[cmd].(*Context); ok {
			panic(kit.Format("%s %s registed by %v", ErrWarn, cmd, s.Name))
		}
		Info.Index[cmd] = s
	})
	kit.If(c.contexts == nil, func() { c.contexts = Contexts{} })
	c.contexts[s.Name] = s
	s.root = c.root
	s.context = c
	s.server = x
	return s
}
func (c *Context) MergeCommands(Commands Commands) *Context {
	for _, cmd := range Commands {
		if cmd.Hand == nil && cmd.RawHand == nil {
			cmd.RawHand = logs.FileLines(2)
		}
	}
	configs := Configs{}
	for k := range Commands {
		configs[k] = &Config{Value: kit.Data()}
	}
	return c.Merge(&Context{Commands: Commands, Configs: configs})
}
func (c *Context) Merge(s *Context) *Context {
	kit.If(c.Commands == nil, func() { c.Commands = Commands{} })
	kit.If(c.Commands[CTX_INIT] == nil, func() { c.Commands[CTX_INIT] = &Command{Hand: func(m *Message, arg ...string) { Info.Load(m) }} })
	kit.If(c.Commands[CTX_EXIT] == nil, func() { c.Commands[CTX_EXIT] = &Command{Hand: func(m *Message, arg ...string) { Info.Save(m) }} })
	merge := func(pre *Command, init bool, key string, cmd *Command, cb Handler) {
		if cb == nil {
			return
		}
		last := pre.Hand
		pre.Hand = func(m *Message, arg ...string) {
			kit.If(init, func() { last(m, arg...) })
			defer kit.If(!init, func() { last(m, arg...) })
			_key, _cmd := m._key, m._cmd
			defer func() { m._key, m._cmd = _key, _cmd }()
			m._key, m._cmd = key, cmd
			cb(m, arg...)
		}
	}
	for key, cmd := range s.Commands {
		if pre, ok := c.Commands[key]; ok && s != c {
			switch key {
			case CTX_INIT:
				merge(pre, true, key, cmd, cmd.Hand)
				continue
			case CTX_EXIT:
				merge(pre, false, key, cmd, cmd.Hand)
				continue
			}
		}
		c.Commands[key] = cmd
		kit.If(cmd.Meta == nil, func() { cmd.Meta = kit.Dict() })
		for sub, action := range cmd.Actions {
			if pre, ok := c.Commands[sub]; ok && s == c {
				kit.Switch(sub,
					CTX_INIT, func() { merge(pre, true, key, cmd, action.Hand) },
					CTX_EXIT, func() { merge(pre, false, key, cmd, action.Hand) },
				)
			}
			if s == c {
				for _, h := range Info.merges {
					switch h := h.(type) {
					case func(c *Context, key string, cmd *Command, sub string, action *Action) Handler:
						merge(c.Commands[CTX_INIT], true, key, cmd, h(c, key, cmd, sub, action))
					case func(c *Context, key string, cmd *Command, sub string, action *Action):
						h(c, key, cmd, sub, action)
					}
				}
			}
			kit.If(sub == SELECT, func() { cmd.Name = kit.Select(action.Name, cmd.Name) })
			kit.If(sub == SELECT, func() { cmd.Help = kit.Select(action.Help, cmd.Help) })
			if help := kit.Split(action.Help, " :："); len(help) > 0 {
				if kit.Value(cmd.Meta, kit.Keys(CTX_TRANS, strings.TrimPrefix(sub, "_")), help[0]); len(help) > 1 {
					kit.Value(cmd.Meta, kit.Keys(CTX_TITLE, sub), help[1])
				}
			}
			kit.Value(cmd.Meta, kit.Keys(CTX_ICONS, sub), action.Icon)
			if action.Hand == nil {
				continue
			}
			kit.If(action.List == nil, func() { action.List = SplitCmd(action.Name, nil) })
			kit.If(len(action.List) > 0, func() { cmd.Meta[sub] = action.List })
		}
		kit.If(strings.HasPrefix(cmd.Name, LIST), func() { cmd.Name = strings.Replace(cmd.Name, LIST, key, 1) })
		kit.If(cmd.List == nil, func() { cmd.List = SplitCmd(cmd.Name, cmd.Actions) })
	}
	kit.If(c.Configs == nil, func() { c.Configs = Configs{} })
	for k, v := range s.Configs {
		if c.Configs[k] == nil || c.Configs[k].Value == nil {
			c.Configs[k] = v
		}
	}
	kit.If(c.Caches == nil, func() { c.Caches = Caches{} })
	return c
}
func (c *Context) Begin(m *Message, arg ...string) *Context {
	kit.If(c.Caches == nil, func() { c.Caches = Caches{} })
	c.Caches[CTX_FOLLOW] = &Cache{Value: c.Name}
	kit.If(c.context != nil && c.context != Index, func() { c.Cap(CTX_FOLLOW, c.context.Prefix(c.Name)) })
	kit.If(c.server != nil, func() { c.server.Begin(m, arg...) })
	return c.Merge(c)
}
func (c *Context) Start(m *Message, arg ...string) {
	m.Log(CTX_START, c.Prefix(), logs.FileLineMeta(2))
	kit.If(c.server != nil, func() { c.server.Start(m, arg...) })
}
func (c *Context) Close(m *Message, arg ...string) {
	m.Log(CTX_CLOSE, c.Prefix(), logs.FileLineMeta(2))
	kit.If(c.server != nil, func() { c.server.Close(m, arg...) })
}

type Message struct {
	time time.Time
	code int

	_data Map
	_meta map[string][]string
	lock  task.Lock

	root    *Message
	message *Message

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

func (m *Message) Time(arg ...string) string {
	t := m.time
	if len(arg) > 0 {
		if d, e := time.ParseDuration(arg[0]); e == nil {
			t, arg = t.Add(d), arg[1:]
		}
	}
	return t.Format(kit.Select(MOD_TIME, arg, 0))
}
func (m *Message) Message() *Message { return m.message }
func (m *Message) Source() *Context  { return m.source }
func (m *Message) Target() *Context  { return m.target }
func (m *Message) _fileline() string {
	switch m.target.Name {
	case MDB, AAA, GDB:
		return m._source
	default:
		return m._target
	}
}
func (m *Message) Spawn(arg ...Any) *Message {
	msg := &Message{time: time.Now(), code: int(m.target.root.ID()),
		_meta: map[string][]string{}, _data: Map{}, message: m, root: m.root,
		_source: logs.FileLine(2), source: m.target, target: m.target, _cmd: m._cmd, _key: m._key, _sub: m._sub,
		W: m.W, R: m.R, O: m.O, I: m.I,
	}
	for _, val := range arg {
		switch val := val.(type) {
		case []byte:
			if m.Warn(json.Unmarshal(val, &msg._meta), string(val)) {
				m.Debug(m.FormatStack(1, 100))
			}
		case Option:
			msg.Option(val.Name, val.Value)
		case Maps:
			kit.For(val, func(k, v string) { msg.Option(k, v) })
		case Map:
			kit.For(kit.KeyValue(nil, "", val), func(k string, v Any) { msg.Option(k, v) })
		case *Context:
			msg.target = val
		case *Command:
			msg._cmd = val
		case string:
			msg._key = val
		case http.ResponseWriter:
			msg.W = val
		case *http.Request:
			msg.R = val
		}
	}
	return msg
}
func (m *Message) Start(key string, arg ...string) *Message {
	return m.Search(key+PT, func(p *Context, s *Context) {
		m.Cmd(ROUTINE, CREATE, kit.Select(m.Prefix(), key), func() { s.Start(m.Spawn(s), arg...) })
	})
}
func (m *Message) Travel(cb Any) *Message {
	target, cmd, key := m.target, m._cmd, m._key
	defer func() { m.target, m._cmd, m._key = target, cmd, key }()
	list := []*Context{m.root.target}
	for i := 0; i < len(list); i++ {
		switch cb := cb.(type) {
		case func(*Context, *Context):
			cb(list[i].context, list[i])
		case func(*Context, *Context, string, *Command):
			m.target = list[i]
			kit.For(kit.SortedKey(list[i].Commands), func(k string) {
				m._cmd, m._key = list[i].Commands[k], k
				cb(list[i].context, list[i], k, list[i].Commands[k])
			})
		case func(*Context, *Context, string, *Config):
			m.target = list[i]
			kit.For(kit.SortedKey(list[i].Configs), func(k string) { cb(list[i].context, list[i], k, list[i].Configs[k]) })
		default:
			m.ErrorNotImplement(cb)
		}
		kit.For(kit.SortedKey(list[i].contexts), func(k string) { list = append(list, list[i].contexts[k]) })
	}
	return m
}
func (m *Message) Search(key string, cb Any) *Message {
	if key == "" {
		return m
	}
	_target, _key, _cmd := m.target, m._key, m._cmd
	defer func() { m.target, m._key, m._cmd = _target, _key, _cmd }()
	p := m.target.root
	if key = strings.TrimPrefix(key, "ice."); key == PT {
		p, key = m.target, ""
	} else if key == ".." {
		p, key = m.target.context, ""
	} else if key == "..." {
		p, key = m.target.root, ""
	} else if strings.Contains(key, PT) {
		ls := strings.Split(key, PT)
		for _, p = range []*Context{m.target.root, m.target, m.source} {
			if p == nil {
				continue
			}
			for _, k := range ls[:len(ls)-1] {
				if p = p.contexts[k]; p == nil {
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
	} else if ctx, ok := Info.Index[key].(*Context); ok {
		p = ctx
	} else {
		p = m.target
	}
	switch cb := cb.(type) {
	case func(key string, cmd *Command):
		if cmd, ok := p.Commands[key]; ok {
			m.target, m._key, m._cmd = p, key, cmd
			cb(key, cmd)
		}
	case func(p *Context, s *Context, key string, cmd *Command):
		for _, p := range []*Context{p, m.target, m.source} {
			for s := p; s != nil; s = s.context {
				if cmd, ok := s.Commands[key]; ok {
					func() {
						_target, _key := m.target, m._key
						defer func() { m.target, m._key = _target, _key }()
						m.target, m._key = s, key
						cb(s.context, s, key, cmd)
					}()
					return m
				}
			}
		}
	case func(p *Context, s *Context, key string, conf *Config):
		for _, p := range []*Context{p, m.target, m.source} {
			for s := p; s != nil; s = s.context {
				if cmd, ok := s.Configs[key]; ok {
					cb(s.context, s, key, cmd)
					return m
				}
			}
		}
	case func(p *Context, s *Context):
		cb(p.context, p)
	default:
		m.ErrorNotImplement(cb)
	}
	return m
}
func (m *Message) Design(action Any, help string, input ...Any) {
	list := kit.List()
	for _, input := range input {
		switch input := input.(type) {
		case string:
			list = append(list, SplitCmd("action "+input, nil)...)
		case Map:
			if kit.Format(input[TYPE]) != "" && kit.Format(input[NAME]) != "" {
				list = append(list, input)
				continue
			}
			kit.For(kit.KeyValue(nil, "", input), func(k string, v Any) {
				list = append(list, kit.Dict(NAME, k, TYPE, TEXT, VALUE, v))
			})
		default:
			m.ErrorNotImplement(input)
		}
	}
	k := kit.Format(action)
	if a, ok := m._cmd.Actions[k]; ok {
		m._cmd.Meta[k], a.List = list, list
		kit.Value(m._cmd.Meta, kit.Keys(CTX_TRANS, k), help)
	}
}
func (m *Message) Actions(key string) *Action   { return m._cmd.Actions[key] }
func (m *Message) Commands(key string) *Command { return m.Target().Commands[kit.Select(m._key, key)] }
