package gdb

import (
	"os"
	"time"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
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
	f.t = time.Tick(kit.Duration(m.Conf(TIMER, kit.Keym(TICK))))
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

			// case <-f.t:
			// m.Cmd(TIMER, ACTION)
		}
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	return true
}

const GDB = "gdb"

var Index = &ice.Context{Name: GDB, Help: "事件模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Cmd(SIGNAL, LISTEN, SIGNAL, "3", kit.MDB_NAME, "退出", ice.CMD, "exit 0")
		m.Cmd(SIGNAL, LISTEN, SIGNAL, "2", kit.MDB_NAME, "重启", ice.CMD, "exit 1")
		m.Load()
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if f, ok := m.Target().Server().(*Frame); ok {
			f.e <- true
		}
		m.Save()
	}},
}}

func init() { ice.Index.Register(Index, &Frame{}, ROUTINE, SIGNAL, EVENT, TIMER) }
