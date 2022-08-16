package ssh

import (
	"encoding/binary"
	"syscall"
	"unsafe"
)

type Winsize struct{ Height, Width, x, y uint16 }

func _ssh_size(fd uintptr, b []byte) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
