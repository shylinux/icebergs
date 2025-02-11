package gdb

import (
	"os"
	"os/signal"
	"syscall"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/file"
)

func _signal_listen(m *ice.Message, s int, arg ...string) {
	if f, ok := m.Target().Server().(*Frame); ok {
		f.listen(m, s, arg...)
	}
}
func _signal_action(m *ice.Message, arg ...string) {
	mdb.HashSelect(m.Spawn(), arg...).Table(func(value ice.Maps) { m.Cmdy(kit.Split(value[ice.CMD])) })
}
func _signal_process(m *ice.Message, p string, s os.Signal) {
	kit.If(p == "", func() { b, _ := file.ReadFile(ice.VAR_LOG_ICE_PID); p = string(b) })
	if p, e := os.FindProcess(kit.Int(kit.Select(kit.Format(os.Getpid()), p))); e == nil {
		p.Signal(s)
	}
}

const (
	PID = "pid"
)
const (
	DEBUG   = "debug"
	ERROR   = "error"
	START   = "start"
	RESTART = "restart"
	STOP    = "stop"
	KILL    = "kill"
)
const SIGNAL = "signal"

func init() {
	Index.MergeCommands(ice.Commands{
		SIGNAL: {Help: "信号量", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { _signal_init(m, arg...) }},
			LISTEN: {Name: "listen signal name cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				_signal_listen(m, kit.Int(m.Option(SIGNAL)), arg...)
			}},
			HAPPEN: {Name: "happen signal", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				_signal_action(m, m.Option(SIGNAL))
			}},
			RESTART: {Name: "restart pid", Hand: func(m *ice.Message, arg ...string) {
				_signal_process(m, m.Option(PID), syscall.SIGINT)
			}},
			STOP: {Name: "stop pid", Hand: func(m *ice.Message, arg ...string) {
				_signal_process(m, m.Option(PID), syscall.SIGQUIT)
			}},
			KILL: {Name: "kill pid signal", Hand: func(m *ice.Message, arg ...string) {
				_signal_process(m, m.Option(PID), syscall.Signal(kit.Int(kit.Select("9", m.Option(SIGNAL)))))
			}},
		}, mdb.HashAction(mdb.SHORT, SIGNAL, mdb.FIELD, "time,signal,name,cmd", mdb.ACTION, HAPPEN), mdb.ClearOnExitHashAction()), Hand: func(m *ice.Message, arg ...string) {
			defer kit.If(len(arg) == 0, func() { m.Action(LISTEN) })
			mdb.HashSelect(m, arg...)
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
