package ice

import (
	"os"
	"strings"
	"sync"
	"time"

	kit "shylinux.com/x/toolkits"
)

type Frame struct {
	wait chan int
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
			list[s] = msg.Spawn(s)
			s.Begin(list[s], arg...)
		}
	})
	return f
}
func (f *Frame) Start(m *Message, arg ...string) bool {
	m.Log(LOG_START, "ice")
	defer m.Cost("start ice")

	m.Cap(CTX_STATUS, CTX_START)
	m.Cap(CTX_STREAM, strings.Split(m.Time(), " ")[1])

	m.Cmdy(INIT, arg)
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
			list[s] = msg.Spawn(s)
			s.Close(list[s], arg...)
		}
	})
	return true
}

var Index = &Context{Name: "ice", Help: "冰山模块", Caches: map[string]*Cache{
	CTX_FOLLOW: {Value: "ice"}, CTX_STREAM: {Value: "shy"}, CTX_STATUS: {Value: CTX_BEGIN},
}, Configs: map[string]*Config{
	HELP: {Value: kit.Data("index", _help)},
}, Commands: map[string]*Command{
	CTX_INIT: {Hand: func(m *Message, c *Context, cmd string, arg ...string) {
		defer m.Cost(CTX_INIT)
		m.root.Travel(func(p *Context, c *Context) {
			if cmd, ok := c.Commands[CTX_INIT]; ok && p != nil {
				c.cmd(m.Spawn(c), cmd, CTX_INIT, arg...)
			}
		})

		m.target.root.wg = &sync.WaitGroup{}
		for _, k := range kit.Split(kit.Select("gdb,log,ssh,mdb")) {
			m.Start(k)
		}
	}},
	INIT: {Name: "init", Help: "启动", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
		m.root.Cmd(CTX_INIT)
		m.Cmd("ssh.source", ETC_INIT_SHY, "init.shy", "启动配置")
		m.Cmdy(arg)
	}},
	HELP: {Name: "help", Help: "帮助", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
		m.Echo(m.Config("index"))
	}},
	EXIT: {Name: "exit", Help: "结束", Hand: func(m *Message, c *Context, cmd string, arg ...string) {
		m.root.Option(EXIT, kit.Select("0", arg, 0))
		m.Cmd("ssh.source", ETC_EXIT_SHY, "exit.shy", "退出配置")
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

		c.server.(*Frame).wait <- kit.Int(m.root.Option(EXIT))
	}},
}}
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

	frame := &Frame{wait: make(chan int, 1)}
	Index.Merge(Index)
	Index.server = frame
	Index.root = Index
	Pulse.root = Pulse

	switch kit.Select("", arg, 0) {
	case "space", "serve":
		if _log_disable = false; frame.Begin(Pulse.Spawn(), arg...).Start(Pulse, arg...) {
			frame.Close(Pulse.Spawn(), arg...)
		}

		os.Exit(<-frame.wait)

	default:
		if Pulse.Cmdy(arg); Pulse.Result() == "" {
			Pulse.Table()
		}
		if strings.TrimSpace(Pulse.Result()) == "" {
			Pulse.Set(MSG_RESULT).Cmdy("cli.system", arg)
		}
		Pulse.Sleep("10ms")
	}

	return Pulse.Result()
}

var _help = `
^_^      欢迎使用冰山框架       ^_^
^_^  Welcome to Icebergs World  ^_^

report: shylinuxc@gmail.com
server: https://shylinux.com
source: https://shylinux.com/x/icebergs
`
