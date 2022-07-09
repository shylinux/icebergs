package ssh

import (
	"net"

	"golang.org/x/crypto/ssh"
	ice "shylinux.com/x/icebergs"
)

type Winsize struct{ Height, Width, x, y uint16 }

func _ssh_size(fd uintptr, b []byte) {
}
func _ssh_sizes(fd uintptr, w, h int) {
}
func _ssh_handle(m *ice.Message, meta ice.Maps, c net.Conn, channel ssh.Channel, requests <-chan *ssh.Request) {
}
