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

func (s *Frame) Begin(m *Message, arg ...string) {
	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		s.root = m.target
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawn(s)
			s.Begin(list[s], arg...)
		}
	})
}
func (s *Frame) Start(m *Message, arg ...string) {
	m.Cmd(INIT, arg)
	kit.For([]string{LOG, GDB, SSH}, func(k string) { m.Sleep30ms().Start(k) })
	m.Sleep30ms().Cmd(arg)
}
func (s *Frame) Close(m *Message, arg ...string) {
	defer conf.Close()
	list := map[*Context]*Message{m.target: m}
	m.Travel(func(p *Context, s *Context) {
		if msg, ok := list[p]; ok && msg != nil {
			list[s] = msg.Spawn(s)
			s.Close(list[s], arg...)
		}
	})
	go func() { os.Exit(kit.Int(Pulse.Sleep300ms().Option(EXIT))) }()
}

const (
	INIT = "init"
	QUIT = "quit"
	EXIT = "exit"
)

var Index = &Context{Name: ICE, Help: "冰山模块", Commands: Commands{
	CTX_INIT: {Hand: func(m *Message, arg ...string) {
		m.Travel(func(p *Context, c *Context) {
			kit.If(p != nil, func() { c._command(m.Spawn(c), c.Commands[CTX_INIT], CTX_INIT, arg...) })
		})
	}},
	CTX_OPEN: {Hand: func(m *Message, arg ...string) {
		m.Travel(func(p *Context, c *Context) {
			kit.If(p != nil, func() { c._command(m.Spawn(c), c.Commands[CTX_OPEN], CTX_OPEN, arg...) })
		})
	}},
	INIT: {Hand: func(m *Message, arg ...string) {
		m.Cmd(kit.Keys(MDB, CTX_INIT))
		m.Cmd(CTX_INIT)
		m.Cmd(SOURCE, ETC_INIT_SHY)
		loadImportant(m)
		m.Cmd(CTX_OPEN)
	}},
	QUIT: {Hand: func(m *Message, arg ...string) {
		m.GoSleep300ms(func() { os.Exit(kit.Int(kit.Select("0", arg, 0))) })
	}},
	EXIT: {Hand: func(m *Message, arg ...string) {
		m.GoSleep300ms(func() {
			m.root.Option(EXIT, kit.Select("0", arg, 0))
			m.Cmd(SOURCE, ETC_EXIT_SHY)
			m.Cmd(CTX_EXIT)
			removeImportant(m)
		})
	}},
	CTX_EXIT: {Hand: func(m *Message, arg ...string) {
		defer m.Target().Close(m.Spawn(), arg...)
		m.Travel(func(p *Context, c *Context) {
			kit.If(p != nil, func() { c._command(m.Spawn(c), c.Commands[CTX_EXIT], CTX_EXIT, arg...) })
		})
	}},
}, server: &Frame{}}

var Pulse = &Message{_meta: map[string][]string{}, _data: Map{}, source: Index, target: Index}

func init() {
	logs.Disable(true)
	switch tz := os.Getenv("TZ"); tz {
	case "", "Asia/Beijing", "Asia/Shanghai":
		time.Local = time.FixedZone(tz, 28800)
	}
	Index.root, Pulse.root, Pulse.time = Index, Pulse, time.Now()
}

func Run(arg ...string) string {
	kit.If(len(arg) == 0 && len(os.Args) > 1, func() { arg = os.Args[1:] })
	if len(arg) == 0 {
		if runtime.GOOS == WINDOWS {
			arg = append(arg, SERVE, START)
		} else {
			arg = append(arg, FOREVER, START)
		}
	} else if arg[0] == FOREVER && arg[1] == START && runtime.GOOS == WINDOWS {
		arg[0] = SERVE
	}
	Pulse.value(MSG_DETAIL, arg...)
	kit.For(kit.Sort(os.Environ()), func(env string) {
		if ls := strings.SplitN(env, EQ, 2); strings.ToLower(ls[0]) == ls[0] && ls[0] != "_" {
			Pulse.Option(ls[0], ls[1])
		}
	})
	kit.If(Pulse._cmd == nil, func() { Pulse._cmd = &Command{RawHand: logs.FileLines(3)} })
	switch Index.Merge(Index).Begin(Pulse, arg...); kit.Select("", arg, 0) {
	case SERVE, SPACE:
		logs.Disable(false)
		Pulse.Go(func() { Index.Start(Pulse, arg...) })
		conf.Wait()
		os.Exit(kit.Int(Pulse.Option(EXIT)))
	default:
		Pulse.Cmdy(INIT).Cmdy(arg)
		kit.If(Pulse.IsErrNotFound(), func() { Pulse.SetAppend().SetResult().Cmdy(SYSTEM, arg) })
		kit.If(strings.TrimSpace(Pulse.Result()) == "" && Pulse.Length() > 0, func() { Pulse.TableEcho() })
		kit.If(Pulse.Result() != "" && !strings.HasSuffix(Pulse.Result(), NL), func() { Pulse.Echo(NL) })
	}
	return Pulse.Result()
}
