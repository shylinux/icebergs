package ice

import (
	"os"
	"runtime"
	"strings"
	"time"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/conf"
	"shylinux.com/x/toolkits/logs"
)

type Frame struct{}

func (s *Frame) Begin(m *Message, arg ...string) Server {
	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		s.root = m.target
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawn(s)
			s.Begin(list[s], arg...)
		}
	})
	return s
}
func (s *Frame) Start(m *Message, arg ...string) bool {
	m.Cap(CTX_STREAM, strings.Split(m.Time(), SP)[1])
	m.Cmd(kit.Keys(MDB, CTX_INIT))
	m.Cmd(kit.Keys(CLI, CTX_INIT))
	m.Cmd(INIT, arg)
	for _, k := range kit.Split(kit.Select("ctx,log,gdb,ssh", os.Getenv(CTX_DAEMON))) {
		m.Start(k)
	}
	m.Cmd(arg)
	return true
}
func (s *Frame) Close(m *Message, arg ...string) bool {
	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawn(s)
			s.Close(list[s], arg...)
		}
	})
	conf.Close()
	go func() { os.Exit(kit.Int(Pulse.Sleep("30ms").Option(EXIT))) }()
	return true
}
func (s *Frame) Spawn(m *Message, c *Context, arg ...string) Server { return &Frame{} }

const (
	INIT = "init"
	HELP = "help"
	EXIT = "exit"
	QUIT = "quit"
)

var Index = &Context{Name: ICE, Help: "冰山模块", Configs: Configs{HELP: {Value: kit.Data(INDEX, Info.Help)}}, Commands: Commands{
	CTX_INIT: {Hand: func(m *Message, arg ...string) {
		m.Travel(func(p *Context, c *Context) {
			if p != nil {
				c._command(m.Spawn(c), c.Commands[CTX_INIT], CTX_INIT, arg...)
			}
		})
	}},
	INIT: {Hand: func(m *Message, arg ...string) {
		m.Cmd(CTX_INIT)
		m.Cmd(SOURCE, ETC_INIT_SHY)
	}},
	HELP: {Hand: func(m *Message, arg ...string) { m.Echo(m.Config(INDEX)) }},
	QUIT: {Hand: func(m *Message, arg ...string) { os.Exit(0) }},
	EXIT: {Hand: func(m *Message, arg ...string) {
		m.root.Option(EXIT, kit.Select("0", arg, 0))
		m.Cmd(SOURCE, ETC_EXIT_SHY)
		m.Cmd(CTX_EXIT)
	}},
	CTX_EXIT: {Hand: func(m *Message, arg ...string) {
		defer m.Target().Close(m.Spawn(), arg...)
		m.Travel(func(p *Context, c *Context) {
			if p != nil {
				c._command(m.Spawn(c), c.Commands[CTX_EXIT], CTX_EXIT, arg...)
			}
		})
	}},
}, server: &Frame{}}
var Pulse = &Message{time: time.Now(), code: 0, meta: map[string][]string{}, data: Map{}, source: Index, target: Index, Hand: true}

func init() { Index.root, Pulse.root = Index, Pulse }

func Run(arg ...string) string {
	if len(arg) == 0 && len(os.Args) > 1 {
		arg = kit.Simple(os.Args[1:], kit.Split(kit.Env(CTX_ARG)))
	}
	if len(arg) == 0 {
		if runtime.GOOS == "windows" {
			arg = append(arg, SERVE, START, DEV, DEV)
		} else {
			arg = append(arg, "forever", START, DEV, DEV)
		}
	}
	Pulse.meta[MSG_DETAIL] = arg
	kit.Fetch(kit.Sort(os.Environ()), func(env string) {
		if ls := strings.SplitN(env, EQ, 2); strings.ToLower(ls[0]) == ls[0] && ls[0] != "_" {
			Pulse.Option(ls[0], ls[1])
		}
	})
	time.Local = time.FixedZone("Beijing", 28800)
	if Pulse.time = time.Now(); Pulse._cmd == nil {
		Pulse._cmd = &Command{RawHand: logs.FileLines(3)}
	}
	switch Index.Merge(Index).Begin(Pulse, arg...); kit.Select("", arg, 0) {
	case SERVE, SPACE:
		if os.Getenv("ctx_log") == "" {
			// os.Stderr.Close()
		}
		if Index.Start(Pulse, arg...) {
			conf.Wait()
			os.Exit(kit.Int(Pulse.Option(EXIT)))
		}
	default:
		if logs.Disable(true); len(arg) == 0 {
			arg = append(arg, HELP)
		}
		if Pulse.Cmdy(INIT).Cmdy(arg); Pulse.IsErrNotFound() {
			Pulse.SetAppend().SetResult().Cmdy(SYSTEM, arg)
		}
		if strings.TrimSpace(Pulse.Result()) == "" {
			Pulse.Table()
		}
		if !strings.HasSuffix(Pulse.Result(), NL) {
			Pulse.Echo(NL)
		}
	}
	return Pulse.Result()
}
