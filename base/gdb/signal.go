package gdb

import (
	"os"
	"os/signal"
	"path"
	"syscall"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	log "shylinux.com/x/toolkits/logs"
)

func _signal_listen(m *ice.Message, s int, arg ...string) {
	if f, ok := m.Target().Server().(*Frame); ok {
		m.Cmdy(mdb.INSERT, SIGNAL, "", mdb.HASH, arg)
		signal.Notify(f.s, syscall.Signal(s))
	}
}
func _signal_action(m *ice.Message, arg ...string) {
	mdb.HashSelect(m.Spawn(), arg...).Table(func(index int, value map[string]string, head []string) {
		m.Cmdy(kit.Split(value[ice.CMD]))
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
	LISTEN = ice.LISTEN
	HAPPEN = ice.HAPPEN
)
const SIGNAL = "signal"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SIGNAL: {Name: SIGNAL, Help: "信号器", Value: kit.Data(
			mdb.SHORT, SIGNAL, mdb.FIELD, "time,signal,name,cmd", nfs.PATH, path.Join(ice.VAR_RUN, "ice.pid"),
		)},
	}, Commands: map[string]*ice.Command{
		SIGNAL: {Name: "signal signal auto listen", Help: "信号器", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				if log.LogDisable {
					return // 禁用日志
				}
				m.Cmd(nfs.SAVE, kit.Select(m.Config(nfs.PATH), m.Conf(cli.RUNTIME, kit.Keys(cli.CONF, cli.CTX_PID))),
					m.Conf(cli.RUNTIME, kit.Keys(cli.HOST, cli.PID)))

				m.Cmd(SIGNAL, LISTEN, SIGNAL, "3", mdb.NAME, "退出", ice.CMD, "exit 0")
				m.Cmd(SIGNAL, LISTEN, SIGNAL, "2", mdb.NAME, "重启", ice.CMD, "exit 1")
			}},
			LISTEN: {Name: "listen signal name cmd", Help: "监听", Hand: func(m *ice.Message, arg ...string) {
				_signal_listen(m, kit.Int(m.Option(SIGNAL)), arg...)
			}},
			HAPPEN: {Name: "happen signal", Help: "触发", Hand: func(m *ice.Message, arg ...string) {
				_signal_action(m, m.Option(SIGNAL))
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Sort(SIGNAL)
			m.PushAction(HAPPEN, mdb.REMOVE)
		}},
	}})
}
