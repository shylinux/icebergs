package ice

import (
	"os"
	"strings"
	"time"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/conf"
	"shylinux.com/x/toolkits/logs"
)

type Frame struct{}

func (f *Frame) Spawn(m *Message, c *Context, arg ...string) Server {
	return &Frame{}
}
func (f *Frame) Begin(m *Message, arg ...string) Server {
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
	m.Cap(CTX_STREAM, strings.Split(m.Time(), SP)[1])
	m.Cmd(kit.Keys(MDB, CTX_INIT))
	m.Cmd("cli.runtime", CTX_INIT)
	m.Cmdy(INIT, arg)

	for _, k := range kit.Split(kit.Select("ctx,log,gdb,ssh", os.Getenv(CTX_DAEMON))) {
		m.Start(k)
	}
	m.Cmd(arg)
	return true
}
func (f *Frame) Close(m *Message, arg ...string) bool {
	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawn(s)
			s.Close(list[s], arg...)
		}
	})
	conf.Close()
	go func() { m.Sleep("1s"); os.Exit(kit.Int(Pulse.Option(EXIT))) }()
	return true
}

const (
	INIT = "init"
	HELP = "help"
	EXIT = "exit"
	QUIT = "quit"
)

var Index = &Context{Name: ICE, Help: "冰山模块", Configs: Configs{
	HELP: {Value: kit.Data(INDEX, Info.Help)},
}, Commands: Commands{
	CTX_INIT: {Hand: func(m *Message, arg ...string) {
		m.root.Travel(func(p *Context, c *Context) {
			if cmd, ok := c.Commands[CTX_INIT]; ok && p != nil {
				c._command(m.Spawn(c), cmd, CTX_INIT, arg...)
			}
		})
	}},
	INIT: {Name: "init", Help: "启动", Hand: func(m *Message, arg ...string) {
		m.root.Cmd(CTX_INIT)
		m.Cmd("source", ETC_INIT_SHY)
	}},
	HELP: {Name: "help", Help: "帮助", Hand: func(m *Message, arg ...string) {
		m.Echo(m.Config(INDEX))
	}},
	QUIT: {Name: "quit", Help: "结束", Hand: func(m *Message, arg ...string) {
		os.Exit(0)
	}},
	EXIT: {Name: "exit", Help: "退出", Hand: func(m *Message, arg ...string) {
		defer m.Target().Close(m.root.Spawn(), arg...)
		m.root.Option(EXIT, kit.Select("0", arg, 0))
		m.Cmd("source", ETC_EXIT_SHY)
		m.root.Cmd(CTX_EXIT)
	}},
	CTX_EXIT: {Hand: func(m *Message, arg ...string) {
		m.root.Travel(func(p *Context, c *Context) {
			if cmd, ok := c.Commands[CTX_EXIT]; ok && p != nil {
				m.TryCatch(m.Spawn(c), true, func(msg *Message) {
					c._command(msg, cmd, CTX_EXIT, arg...)
				})
			}
		})
	}},
}, server: &Frame{}}
var Pulse = &Message{time: time.Now(), code: 0,
	meta: map[string][]string{}, data: Map{},
	source: Index, target: Index, Hand: true,
}

func init() { Index.root, Pulse.root = Index, Pulse }

func Run(arg ...string) string {
	if len(arg) == 0 { // 进程参数
		arg = kit.Simple(arg, os.Args[1:], kit.Split(os.Getenv(CTX_ARG)))
	}

	Pulse.meta[MSG_DETAIL] = arg
	switch Index.Merge(Index).Begin(Pulse.Spawn(), arg...); kit.Select("", arg, 0) {
	case "serve", "space": // 启动服务
		if Index.Start(Pulse, arg...) {
			conf.Wait()
			println()
			os.Exit(kit.Int(Pulse.Option(EXIT)))
		}
	default: // 执行命令
		if logs.Disable(true); len(arg) == 0 {
			arg = append(arg, HELP)
		}
		Pulse.Cmd(INIT)
		if Pulse.Cmdy(arg); strings.TrimSpace(Pulse.Result()) == "" {
			Pulse.Table()
		}
	}

	if !strings.HasSuffix(Pulse.Result(), NL) {
		Pulse.Echo(NL)
	}
	return Pulse.Result()
}
