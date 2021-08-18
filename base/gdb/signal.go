package gdb

import (
	"os"
	"os/signal"
	"path"
	"syscall"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _signal_listen(m *ice.Message, s int, arg ...string) {
	if f, ok := m.Target().Server().(*Frame); ok {
		m.Cmdy(mdb.INSERT, SIGNAL, "", mdb.HASH, arg)
		signal.Notify(f.s, syscall.Signal(s))
	}
}
func _signal_action(m *ice.Message, s int) {
	m.Option(mdb.FIELDS, "time,signal,name,cmd")
	msg := m.Cmd(mdb.SELECT, SIGNAL, "", mdb.HASH, SIGNAL, s)
	msg.Table(func(index int, value map[string]string, head []string) {
		m.Cmdy(kit.Split(value[cli.CMD]))
	})
}

func SignalNotify(m *ice.Message, sig int, cb func()) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.Signal(sig))
	m.Go(func() {
		for {
			if _, ok := <-ch; ok {
				cb()
			}
		}
	})
}

const (
	LISTEN = "listen"
	ACTION = "action"
)
const SIGNAL = "signal"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SIGNAL: {Name: SIGNAL, Help: "信号器", Value: kit.Data(
				kit.MDB_PATH, path.Join(ice.VAR_RUN, "ice.pid"), kit.MDB_SHORT, SIGNAL,
			)},
		},
		Commands: map[string]*ice.Command{
			SIGNAL: {Name: "signal signal auto listen", Help: "信号器", Action: map[string]*ice.Action{
				LISTEN: {Name: "listen signal name cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
					_signal_listen(m, kit.Int(m.Option(SIGNAL)), arg...)
				}},
				ACTION: {Name: "action signal", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
					_signal_action(m, kit.Int(m.Option(SIGNAL)))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SIGNAL, "", mdb.HASH, SIGNAL, m.Option(SIGNAL))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,signal,name,cmd")
				m.Cmdy(mdb.SELECT, SIGNAL, "", mdb.HASH, SIGNAL, arg)
				m.PushAction(ACTION, mdb.REMOVE)
				m.Sort(SIGNAL)
			}},
		},
	})
}
