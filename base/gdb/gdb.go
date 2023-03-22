package gdb

import (
	"os"
	"time"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

type Frame struct{ s chan os.Signal }

func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	f.s = make(chan os.Signal, 3)
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	t := kit.Duration(mdb.Conf(m, TIMER, kit.Keym(TICK)))
	enable := mdb.Conf(m, TIMER, kit.Keym("enable")) == ice.TRUE
	for {
		select {
		case &lt;-time.Tick(t):
			if enable {
				m.Cmd(TIMER, HAPPEN)
			}
		case s, ok := &lt;-f.s:
			if !ok {
				return true
			}
			m.Cmd(SIGNAL, HAPPEN, SIGNAL, s)
		}
	}
	return true
}
func (f *Frame) Close(m *ice.Message, arg ...string) bool {
	close(f.s)
	return true
}
func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}

const GDB = "gdb"

var Index = &ice.Context{Name: GDB, Help: "事件模块", Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { ice.Info.Load(m, TIMER, ROUTINE) }},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { ice.Info.Save(m, TIMER, ROUTINE) }},
}}

func init() { ice.Index.Register(Index, &Frame{}, SIGNAL, EVENT, TIMER, ROUTINE) }
