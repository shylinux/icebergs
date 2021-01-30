package ssh

import (
	ice "github.com/shylinux/icebergs"
	"golang.org/x/crypto/ssh"

	"encoding/binary"
	"net"
	"syscall"
	"unsafe"
)

func _ssh_size(fd uintptr, b []byte) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])

	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
func _ssh_handle(m *ice.Message, meta map[string]string, c net.Conn, channel ssh.Channel, requests <-chan *ssh.Request) {
}
