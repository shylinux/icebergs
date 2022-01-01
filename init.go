package ice

import (
	"os"
	"strings"
	"sync"
	"time"

	kit "shylinux.com/x/toolkits"
	log "shylinux.com/x/toolkits/logs"
)

type Frame struct{}

func (f *Frame) Spawn(m *Message, c *Context, arg ...string) Server {
	return &Frame{}
}
func (f *Frame) Begin(m *Message, arg ...string) Server {
	defer m.Cost(LOG_BEGIN, ICE)

	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		s.root = m.target
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawn(s)
			s.Begin(list[s], arg...)
		}
	})
	return f
}
func (f *Frame) Start(m *Message, arg ...string) bool {
	defer m.Cost(LOG_START, ICE)

	m.Cap(CTX_STREAM, strings.Split(m.Time(), SP)[1])
	m.Cmdy(INIT, arg)

	for _, k := range kit.Split("log,gdb,ssh") {
		m.Start(k)
	}
	m.Cmdy(arg)
	return true
}
func (f *Frame) Close(m *Message, arg ...string) bool {
	defer m.Cost(LOG_CLOSE, ICE)

	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawn(s)
			s.Close(list[s], arg...)
		}
	})
	return true
}

var Index = &Context{Name: "ice", Help: "冰山模块", Configs: map[string]*Config{
	HELP: {Value: kit.Data("index", Info.Help)},
}, Commands: map[string]*Command{
	CTX_INIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
		defer m.Cost(CTX_INIT)
		m.root.Travel(func(p *Context, c *Context) {
			if cmd, ok := c.Commands[CTX_INIT]; ok && p != nil {
				c.cmd(m.Spawn(c), cmd, CTX_INIT, arg...)
			}
		})
	}},
	INIT: {Name: "init", Help: "启动", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
		m.root.Cmd(CTX_INIT)
		m.Cmd("source", ETC_INIT_SHY)
	}},
	HELP: {Name: "help", Help: "帮助", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
		m.Echo(m.Config("index"))
	}},
	EXIT: {Name: "exit", Help: "结束", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
		m.root.Option(EXIT, kit.Select("0", arg, 0))
		defer c.Close(m.root.Spawn(), arg...)

		m.Cmd("source", ETC_EXIT_SHY)
		m.root.Cmd(CTX_EXIT)
	}},
	CTX_EXIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
		defer m.Cost(CTX_EXIT)
		m.root.Travel(func(p *Context, c *Context) {
			if cmd, ok := c.Commands[CTX_EXIT]; ok && p != nil {
				m.TryCatch(m.Spawn(c), true, func(msg *Message) {
					c.cmd(msg, cmd, CTX_EXIT, arg...)
				})
			}
		})
	}},
}, server: &Frame{}, wg: &sync.WaitGroup{}}
var Pulse = &Message{
	time: time.Now(), code: 0,
	meta: map[string][]string{},
	data: map[string]interface{}{},

	source: Index, target: Index, Hand: true,
}

func Run(arg ...string) string {
	if len(arg) == 0 {
		arg = os.Args[1:]
	}
	if len(arg) == 0 {
		arg = append(arg, HELP)
	}
	if os.Getenv("ctx_arg") != "" {
		arg = append(arg, kit.Split(os.Getenv("ctx_arg"))...)
	}

	Index.root, Pulse.root = Index, Pulse

	switch Index.Merge(Index).Begin(Pulse.Spawn(), arg...); kit.Select("", arg, 0) {
	case "serve", "space":
		if log.LogDisable = false; Index.Start(Pulse, arg...) {
			Pulse.TryCatch(Pulse, true, func(Pulse *Message) { Index.wg.Wait() })
			os.Exit(kit.Int(Pulse.Option(EXIT)))
		}
	default:
		if Pulse.Cmdy(arg); Pulse.Result() == "" {
			Pulse.Table()
		}
		Pulse.Sleep30ms()
	}

	return Pulse.Result()
}
