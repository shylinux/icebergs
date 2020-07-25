package ice

import (
	kit "github.com/shylinux/toolkits"

	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Frame struct {
	code int
}

func (f *Frame) Spawn(m *Message, c *Context, arg ...string) Server {
	return &Frame{}
}
func (f *Frame) Begin(m *Message, arg ...string) Server {
	m.Log(LOG_BEGIN, "ice")
	defer m.Cost("begin ice")

	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		s.root = m.target
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawns(s)
			s.Begin(list[s], arg...)
		}
	})
	return f
}
func (f *Frame) Start(m *Message, arg ...string) bool {
	m.Log(LOG_START, "ice")
	defer m.Cost("start ice")

	m.Cap(CTX_STATUS, "start")
	m.Cap(CTX_STREAM, strings.Split(m.Time(), " ")[1])

	m.Cmdy("init", arg)
	return true
}
func (f *Frame) Close(m *Message, arg ...string) bool {
	m.TryCatch(m, true, func(m *Message) {
		m.target.wg.Wait()
	})

	m.Log(LOG_CLOSE, "ice")
	defer m.Cost("close ice")

	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawns(s)
			s.Close(list[s], arg...)
		}
	})
	return true
}

var wait = make(chan bool, 1)
var Index = &Context{Name: "ice", Help: "冰山模块",
	Caches: map[string]*Cache{
		CTX_FOLLOW: {Value: ""},
		CTX_STREAM: {Value: "shy"},
		CTX_STATUS: {Value: "begin"},
	},
	Configs: map[string]*Config{
		"help": {Value: map[string]interface{}{
			"index": []interface{}{
				"^_^      欢迎使用冰山框架       ^_^",
				"^_^  Welcome to Icebergs world  ^_^",
				"",
				"Meet: shylinuxc@gmail.com",
				"More: https://shylinux.com",
				"More: https://github.com/shylinux/icebergs",
				"",
			},
		}},
		"task": {Value: kit.Dict(
			kit.MDB_STORE, "var/data",
			kit.MDB_LIMIT, "110",
			kit.MDB_LEAST, "100",
		)},
	},
	Commands: map[string]*Command{
		CTX_INIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			defer m.Cost("_init ice")
			m.Travel(func(p *Context, c *Context) {
				if cmd, ok := c.Commands[CTX_INIT]; ok && p != nil {
					c.cmd(m.Spawns(c), cmd, CTX_INIT, arg...)
				}
			})
		}},
		"init": {Name: "init", Help: "启动", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.root.Cmd(CTX_INIT)

			m.target.root.wg = &sync.WaitGroup{}
			for _, k := range kit.Split(kit.Select("gdb,log,ssh,ctx", os.Getenv("ctx_mod"))) {
				m.Start(k)
			}

			m.Cmd("ssh.source", "etc/init.shy", "init.shy", "启动配置")
			m.Cmdy(arg)
		}},
		"help": {Name: "help", Help: "帮助", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Echo(strings.Join(kit.Simple(m.Confv("help", "index")), "\n"))
		}},
		"name": {Name: "name", Help: "命名", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			for k, v := range names {
				m.Push("key", k)
				switch v := v.(type) {
				case *Context:
					m.Push("value", v.Name)
				default:
					m.Push("value", "")
				}
			}
			m.Sort("key")
		}},
		"exit": {Name: "exit", Help: "结束", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.root.target.server.(*Frame).code = kit.Int(kit.Select("0", arg, 0))
			m.Cmd("ssh.source", "etc/exit.shy", "exit.shy", "退出配置")

			m.root.Cmd(CTX_EXIT)
		}},
		CTX_EXIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			defer m.Cost(CTX_EXIT)
			m.root.Travel(func(p *Context, c *Context) {
				if cmd, ok := c.Commands[CTX_EXIT]; ok && p != nil {
					m.TryCatch(m.Spawns(c), true, func(msg *Message) {
						c.cmd(msg, cmd, CTX_EXIT, arg...)
					})
				}
			})
			wait <- true
		}},
	},
}

var Pulse = &Message{
	time: time.Now(), code: 0,
	meta: map[string][]string{},
	data: map[string]interface{}{},

	source: Index, target: Index, Hand: true,
	messages: []*Message{}, message: nil, root: nil,
}

var Log func(*Message, string, string)
var Loop func()

func Run(arg ...string) string {
	if len(arg) == 0 {
		arg = os.Args[1:]
	}
	if len(arg) == 0 {
		arg = append(arg, "web.space", "connect", "self")
	}

	frame := &Frame{}
	Index.root = Index
	Index.server = frame

	Pulse.root = Pulse
	Pulse.Option("cache.limit", "30")
	Pulse.Option("begin_time", Pulse.Time())

	if frame.Begin(Pulse.Spawns(), arg...).Start(Pulse, arg...) {
		if Loop != nil {
			Loop()
		}
		frame.Close(Pulse.Spawns(), arg...)
	}

	if Pulse.Result() == "" {
		Pulse.Table(nil)
	}
	fmt.Printf(Pulse.Result())
	<-wait
	os.Exit(frame.code)
	return ""
}

var names = map[string]interface{}{}

var ErrNameExists = "name already exists:"

type Error struct {
	Arg      []interface{}
	FileLine string
}

func NewError(n int, arg ...interface{}) *Error {
	return &Error{Arg: arg, FileLine: kit.FileLine(n, 3)}
}
func (e *Error) Error() string {
	return e.FileLine + " " + strings.Join(kit.Simple(e.Arg), " ")
}

func Name(name string, value interface{}) string {
	if s, ok := names[name]; ok {
		last := ""
		switch s := s.(type) {
		case *Context:
			last = s.Name
		}
		panic(NewError(4, ErrNameExists, name, "last:", last))
	}
	names[name] = value
	return name
}
