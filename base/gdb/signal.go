package gdb

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"os/signal"
	"syscall"
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
		m.Cmdy(kit.Split(value[kit.SSH_CMD]))
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
				kit.MDB_PATH, "var/run/ice.pid", kit.MDB_SHORT, SIGNAL,
			)},
		},
		Commands: map[string]*ice.Command{
			SIGNAL: {Name: "signal auto 监听", Help: "信号器", Action: map[string]*ice.Action{
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
				m.Option(mdb.FIELDS, "time,signal,name,cmd")
				m.Cmdy(mdb.SELECT, SIGNAL, "", mdb.HASH, SIGNAL, arg)
				m.PushAction(ACTION, mdb.REMOVE)
				m.Sort(SIGNAL)
			}},
		},
	}, nil)
}
