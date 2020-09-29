package ice

import (
	kit "github.com/shylinux/toolkits"

	"os"
	"strings"
	"sync"
	"time"
)

type Frame struct{ code int }

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
				"^_^  Welcome to Icebergs World  ^_^",
				"",
				"More: shylinuxc@gmail.com",
				"More: https://shylinux.com",
				"More: https://github.com/shylinux/icebergs",
				"",
			},
		}},
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
			for _, k := range kit.Split(kit.Select("gdb,log,ssh,mdb")) {
				m.Start(k)
			}
			m.Cmd("ssh.source", "etc/init.shy", "init.shy", "启动配置")
			m.Cmdy(arg)
		}},
		"help": {Name: "help", Help: "帮助", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
			m.Echo(strings.Join(kit.Simple(m.Confv("help", "index")), "\n"))
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
}
var wait = make(chan bool, 1)

func Run(arg ...string) string {
	if len(arg) == 0 {
		arg = os.Args[1:]
	}
	if len(arg) == 0 {
		arg = append(arg, "help")
	}

	frame := &Frame{}
	Index.root = Index
	Index.server = frame

	Pulse.root = Pulse
	Pulse.Option("cache.limit", "30")
	Pulse.Option("begin_time", Pulse.Time())

	switch kit.Select("", arg, 0) {
	case "space", "serve":
		if _log_disable = false; frame.Begin(Pulse.Spawns(), arg...).Start(Pulse, arg...) {
			frame.Close(Pulse.Spawns(), arg...)
		}

		<-wait
		os.Exit(frame.code)
	default:
		if m := Pulse.Cmdy(arg); m.Result() == "" {
			m.Table()
		}
	}

	return Pulse.Result()
}

var names = map[string]interface{}{}

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
