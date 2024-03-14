package gdb

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

func _signal_init(m *ice.Message, arg ...string) {
	_signal_listen(m, 1, mdb.NAME, START, ice.CMD, "runtime")
	_signal_listen(m, 2, mdb.NAME, RESTART, ice.CMD, "exit 1")
	_signal_listen(m, 3, mdb.NAME, STOP, ice.CMD, "exit 0")
}
func SignalProcess(m *ice.Message, pid string) bool {
	return false
}
