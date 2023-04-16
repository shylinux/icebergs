//go:build darwin
// +build darwin

package xterm

import (
	"errors"
	"os"
	"syscall"
	"unsafe"

	kit "shylinux.com/x/toolkits"
)

func Open() (pty, tty *os.File, err error) {
	pFD, err := syscall.Open("/dev/ptmx", syscall.O_RDWR|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}
	p := os.NewFile(uintptr(pFD), "/dev/ptmx")
	defer func() { kit.If(err != nil, func() { _ = p.Close() }) }()
	sname, err := ptsname(p)
	if err != nil {
		return nil, nil, err
	}
	if err := grantpt(p); err != nil {
		return nil, nil, err
	}
	if err := unlockpt(p); err != nil {
		return nil, nil, err
	}
	t, err := os.OpenFile(sname, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, err
	}
	return p, t, nil
}
func ptsname(f *os.File) (string, error) {
	n := make([]byte, _IOC_PARM_LEN(syscall.TIOCPTYGNAME))
	err := ioctl(f.Fd(), syscall.TIOCPTYGNAME, uintptr(unsafe.Pointer(&n[0])))
	if err != nil {
		return "", err
	}
	for i, c := range n {
		if c == 0 {
			return string(n[:i]), nil
		}
	}
	return "", errors.New("TIOCPTYGNAME string not NUL-terminated")
}
func grantpt(f *os.File) error  { return ioctl(f.Fd(), syscall.TIOCPTYGRANT, 0) }
func unlockpt(f *os.File) error { return ioctl(f.Fd(), syscall.TIOCPTYUNLK, 0) }

const (
	_IOC_VOID    uintptr = 0x20000000
	_IOC_OUT     uintptr = 0x40000000
	_IOC_IN      uintptr = 0x80000000
	_IOC_IN_OUT  uintptr = _IOC_OUT | _IOC_IN
	_IOC_DIRMASK         = _IOC_VOID | _IOC_OUT | _IOC_IN

	_IOC_PARAM_SHIFT = 13
	_IOC_PARAM_MASK  = (1 << _IOC_PARAM_SHIFT) - 1
)

func _IOC_PARM_LEN(ioctl uintptr) uintptr { return (ioctl >> 16) & _IOC_PARAM_MASK }
