package gdb

import (
	"os"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

type Frame struct {
	s chan os.Signal
}

func (f *Frame) Spawn(m *ice.Message, c *ice.Context, arg ...string) ice.Server {
	return &Frame{}
}
func (f *Frame) Begin(m *ice.Message, arg ...string) ice.Server {
	list := []string{}
	for k := range ice.Info.File {
		if strings.HasPrefix(k, ice.Info.Make.Path+ice.PS) {
			list = append(list, k)
		}
	}
	for _, k := range list {
		ice.Info.File["/require/"+strings.TrimPrefix(k, ice.Info.Make.Path+ice.PS)] = ice.Info.File[k]
		delete(ice.Info.File, k)
	}

	f.s = make(chan os.Signal, 3)
	return f
}
func (f *Frame) Start(m *ice.Message, arg ...string) bool {
	t := kit.Duration(m.Conf(TIMER, kit.Keym(TICK)))
	enable := m.Conf(TIMER, kit.Keym("enable")) == ice.TRUE

	for {
		select {
		case <-time.Tick(t):
			if enable {
				m.Cmd(TIMER, HAPPEN)
			}

		case s, ok := <-f.s:
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

const GDB = "gdb"

var Index = &ice.Context{Name: GDB, Help: "事件模块", Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { ice.Info.Load(m, TIMER, ROUTINE) }},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { ice.Info.Save(m, TIMER, ROUTINE) }},
}}

func init() { ice.Index.Register(Index, &Frame{}, SIGNAL, TIMER, EVENT, ROUTINE) }
