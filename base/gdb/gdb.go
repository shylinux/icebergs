package gdb

import (
	"os"
	"time"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

type Frame struct {
	s chan os.Signal
	t time.Duration
	e chan bool
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	f.s = make(chan os.Signal, ice.MOD_CHAN)
	f.e = make(chan bool, 1)
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	f.t = kit.Duration(m.Conf(TIMER, kit.Keym(TICK)))
	for {
		select {
		case <-f.e:
			return true

		case <-time.Tick(f.t):
			// m.Cmd(TIMER, ACTION)

		case s := <-f.s:
			m.Cmd(SIGNAL, ACTION, ACTION, SIGNAL, s)
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
		m.Load(TIMER)
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if f, ok := m.Target().Server().(*Frame); ok {
			f.e <- true
		}
		m.Save(TIMER)
	}},
}}

func init() { ice.Index.Register(Index, &Frame{}, ROUTINE, SIGNAL, EVENT, TIMER) }
