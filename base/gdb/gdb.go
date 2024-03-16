package gdb

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

type Frame struct{ s chan os.Signal }

func (f *Frame) Begin(m *ice.Message, arg ...string) {
	f.s = make(chan os.Signal, 10)
}
func (f *Frame) Start(m *ice.Message, arg ...string) {
	if f, p, e := logs.CreateFile(ice.VAR_LOG_ICE_PID); e == nil {
		m.Logs("save", "file", p, PID, os.Getpid())
		fmt.Fprint(f, os.Getpid())
		f.Close()
	}
	t := time.NewTicker(kit.Duration(mdb.Conf(m, TIMER, kit.Keym(TICK))))
	for {
		select {
		case <-t.C:
			m.Options(ice.LOG_DISABLE, ice.TRUE).Cmd(TIMER, HAPPEN)
		case s, ok := <-f.s:
			if !ok {
				return
			}
			m.Cmd(SIGNAL, HAPPEN, SIGNAL, s)
		}
	}
}
func (f *Frame) Close(m *ice.Message, arg ...string) {
	close(f.s)
}
func (f *Frame) listen(m *ice.Message, s int, arg ...string) {
	signal.Notify(f.s, syscall.Signal(s))
	mdb.HashCreate(m, SIGNAL, s, arg)
}

const GDB = "gdb"

var Index = &ice.Context{Name: GDB, Help: "事件模块"}

func Prefix(arg ...string) string { return kit.Keys(GDB, arg) }

func init() { ice.Index.Register(Index, &Frame{}, SIGNAL, EVENT, TIMER, ROUTINE) }
