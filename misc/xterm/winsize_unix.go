//go:build !windows
// +build !windows

package xterm

import (
	"os"
	"syscall"
	"unsafe"
)

func Setsize(t *os.File, ws *Winsize) error {
	return ioctl(t.Fd(), syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(ws)))
}

const (
	TIOCGWINSZ = syscall.TIOCGWINSZ
	TIOCSWINSZ = syscall.TIOCSWINSZ
)

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if e != 0 {
		return e
	}
	return nil
}
