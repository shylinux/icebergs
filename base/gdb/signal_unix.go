//go:build !windows

package gdb

import (
	"os"
	"syscall"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _signal_init(m *ice.Message, arg ...string) {
	_signal_listen(m, 1, mdb.NAME, START, ice.CMD, "runtime")
	_signal_listen(m, 2, mdb.NAME, RESTART, ice.CMD, "exit 1")
	_signal_listen(m, 3, mdb.NAME, STOP, ice.CMD, "exit 0")
	_signal_listen(m, int(syscall.SIGUSR1), mdb.NAME, "info", ice.CMD, "runtime")
}
func SignalProcess(m *ice.Message, pid string) bool {
	if proc, err := os.FindProcess(kit.Int(pid)); err == nil && proc.Signal(syscall.SIGUSR1) == nil {
		return true
	}
	return false
}
