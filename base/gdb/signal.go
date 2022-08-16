package gdb

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
	"shylinux.com/x/toolkits/logs"
)

func _signal_listen(m *ice.Message, s int, arg ...string) {
	if f, ok := m.Target().Server().(*Frame); ok {
		signal.Notify(f.s, syscall.Signal(s))
		mdb.HashCreate(m, SIGNAL, s, arg)
	}
}
func _signal_action(m *ice.Message, arg ...string) {
	mdb.HashSelect(m.Spawn(), arg...).Tables(func(value ice.Maps) {
		m.Cmdy(kit.Split(value[ice.CMD]))
	})
}
func _signal_process(m *ice.Message, p string, s os.Signal) {
	if p == "" {
		b, _ := file.ReadFile(ice.Info.PidPath)
		p = string(b)
	}
	if p == "" {
		p = kit.Format(os.Getpid())
	}
	if p, e := os.FindProcess(kit.Int(p)); e == nil {
		p.Signal(s)
	}
}

const (
	PID = "pid"
)
const (
	LISTEN = "listen"
	HAPPEN = "happen"

	START   = "start"
	RESTART = "restart"
	STOP    = "stop"
	ERROR   = "error"
	KILL    = "kill"
)
const SIGNAL = "signal"

func init() {
	Index.MergeCommands(ice.Commands{
		SIGNAL: {Name: "signal signal auto listen", Help: "信号器", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				_signal_listen(m, 2, mdb.NAME, "重启", ice.CMD, "exit 1")
				_signal_listen(m, 3, mdb.NAME, "退出", ice.CMD, "exit 0")
				if f, p, e := logs.CreateFile(ice.Info.PidPath); e == nil {
					defer f.Close()
					fmt.Fprint(f, os.Getpid())
					m.Logs(mdb.CREATE, PID, p)
				}
			}},
			LISTEN: {Name: "listen signal name cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				_signal_listen(m, kit.Int(m.Option(SIGNAL)), arg...)
			}},
			HAPPEN: {Name: "happen signal", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				_signal_action(m, m.Option(SIGNAL))
			}},
			RESTART: {Name: "restart pid", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				_signal_process(m, m.Option(PID), syscall.SIGINT)
			}},
			STOP: {Name: "stop pid", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				_signal_process(m, m.Option(PID), syscall.SIGQUIT)
			}},
			KILL: {Name: "kill pid signal", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				_signal_process(m, m.Option(PID), syscall.Signal(kit.Int(kit.Select("9", m.Option(SIGNAL)))))
			}},
		}, mdb.HashAction(mdb.SHORT, SIGNAL, mdb.FIELD, "time,signal,name,cmd", mdb.ACTION, HAPPEN)), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Sort(SIGNAL)
		}},
	})
}

func SignalNotify(m *ice.Message, sig syscall.Signal, cb func()) {
	ch := make(chan os.Signal)
	signal.Notify(ch, sig)
	m.Go(func() {
		for {
			if _, ok := <-ch; ok {
				cb()
			}
		}
	})
}
