//go:build linux
// +build linux

package xterm

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"

	kit "shylinux.com/x/toolkits"
)

func Open() (*os.File, *os.File, error) {
	p, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	defer func() { kit.If(err != nil, func() { _ = p.Close() }) }()
	sname, err := ptsname(p)
	if err != nil {
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
	var n uint32
	err := ioctl(f.Fd(), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&n)))
	if err != nil {
		return "", err
	}
	return "/dev/pts/" + strconv.Itoa(int(n)), nil
}

func unlockpt(f *os.File) error {
	var u int32
	return ioctl(f.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&u)))
}
