package gdb

import (
	"os"
	"time"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

type Frame struct {
	t <-chan time.Time
	s chan os.Signal
	e chan bool
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	f.t = time.Tick(kit.Duration(m.Conf(TIMER, kit.Keym("tick"))))
	f.s = make(chan os.Signal, ice.MOD_CHAN)
	f.e = make(chan bool, 1)
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	for {
		select {
		case <-f.e:
			return true

		case s := <-f.s:
			m.Cmd(SIGNAL, ACTION, SIGNAL, s)

		case <-f.t:
			_timer_action(m.Spawn())
		}
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

const (
	BUILD = "build"
	SPAWN = "spawn"
	START = "start"
	STOP  = "stop"

	STATUS  = "status"
	RESTART = "restart"
	RELOAD  = "reload"

	BENCH = "bench"
	PPROF = "pprof"
	BEGIN = "begin"
	END   = "end"
)

const GDB = "gdb"

var Index = &ice.Context{Name: GDB, Help: "事件模块",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(nfs.SAVE, kit.Select(m.Conf(SIGNAL, kit.META_PATH), m.Conf(cli.RUNTIME, kit.Keys(cli.CONF, cli.CTX_PID))),
				m.Conf(cli.RUNTIME, kit.Keys(cli.HOST, "pid")))

			m.Cmd(SIGNAL, LISTEN, SIGNAL, "3", kit.MDB_NAME, "退出", kit.SSH_CMD, "exit 0")
			m.Cmd(SIGNAL, LISTEN, SIGNAL, "2", kit.MDB_NAME, "重启", kit.SSH_CMD, "exit 1")
			m.Load()
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if f, ok := m.Target().Server().(*Frame); ok {
				f.e <- true
			}
			m.Save(TIMER)
		}},
	},
}

func init() {
	ice.Index.Register(Index, &Frame{}, ROUTINE, SIGNAL, EVENT, TIMER)
}
