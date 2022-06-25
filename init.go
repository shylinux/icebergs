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

	for _, k := range kit.Split(Getenv("ctx_daemon")) {
		m.Start(k)
	}
	m.Cmd(arg)
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

var Index = &Context{Name: ICE, Help: "冰山模块", Configs: map[string]*Config{
	HELP: {Value: kit.Data(INDEX, Info.Help)},
}, Commands: map[string]*Command{
	CTX_INIT: {Hand: func(m *Message, arg ...string) {
		defer m.Cost(CTX_INIT)
		m.root.Travel(func(p *Context, c *Context) {
			if cmd, ok := c.Commands[CTX_INIT]; ok && p != nil {
				c._command(m.Spawn(c), cmd, CTX_INIT, arg...)
			}
		})
	}},
	INIT: {Name: "init", Help: "启动", Hand: func(m *Message, arg ...string) {
		m.root.Cmd(CTX_INIT)
		m.Cmd(SOURCE, ETC_INIT_SHY)
	}},
	HELP: {Name: "help", Help: "帮助", Hand: func(m *Message, arg ...string) {
		m.Echo(m.Config(INDEX))
	}},
	EXIT: {Name: "exit", Help: "结束", Hand: func(m *Message, arg ...string) {
		m.root.Option(EXIT, kit.Select("0", arg, 0))
		defer m.Target().Close(m.root.Spawn(), arg...)

		m.Cmd(SOURCE, ETC_EXIT_SHY)
		m.root.Cmd(CTX_EXIT)
	}},
	CTX_EXIT: {Hand: func(m *Message, arg ...string) {
		defer m.Cost(CTX_EXIT)
		m.Option("cmd_dir", "")
		m.Option("dir_root", "")
		m.root.Travel(func(p *Context, c *Context) {
			if cmd, ok := c.Commands[CTX_EXIT]; ok && p != nil {
				m.TryCatch(m.Spawn(c), true, func(msg *Message) {
					c._command(msg, cmd, CTX_EXIT, arg...)
				})
			}
		})
		_exit <- kit.Int(m.Option(EXIT))
	}},
}, server: &Frame{}, wg: &sync.WaitGroup{}}
var Pulse = &Message{time: time.Now(), code: 0,
	meta: map[string][]string{}, data: Map{},
	source: Index, target: Index, Hand: true,
}

var _exit = make(chan int, 1)

func init() { Index.root, Pulse.root = Index, Pulse }

func Run(arg ...string) string {
	list := []string{}
	for k := range Info.File {
		if strings.HasPrefix(k, Info.Make.Path+PS) {
			list = append(list, k)
		}
	}
	for _, k := range list {
		Info.File["/require/"+strings.TrimPrefix(k, Info.Make.Path+PS)] = Info.File[k]
		delete(Info.File, k)
	}

	if len(arg) == 0 { // 进程参数
		if arg = append(arg, os.Args[1:]...); kit.Env("ctx_arg") != "" {
			arg = append(arg, kit.Split(kit.Env("ctx_arg"))...)
		}
	}

	Pulse.meta[MSG_DETAIL] = arg
	switch Index.Merge(Index).Begin(Pulse.Spawn(), arg...); kit.Select("", arg, 0) {
	case SERVE, SPACE: // 启动服务
		switch strings.Split(os.Getenv("TERM"), "-")[0] {
		case "xterm", "screen":
			Info.Colors = true
		default:
			Info.Colors = false
		}
		if log.LogDisable = false; Index.Start(Pulse, arg...) {
			Pulse.TryCatch(Pulse, true, func(Pulse *Message) { Index.wg.Wait() })
			os.Exit(<-_exit)
		}
	default: // 执行命令
		if len(arg) == 0 {
			arg = append(arg, HELP)
		}

		Pulse.Cmd(INIT)
		if Pulse.Cmdy(arg); strings.TrimSpace(Pulse.Result()) == "" {
			Pulse.Table()
		}
		Pulse.Sleep30ms()
	}

	if !strings.HasSuffix(Pulse.Result(), NL) {
		Pulse.Echo(NL)
	}
	return Pulse.Result()
}
