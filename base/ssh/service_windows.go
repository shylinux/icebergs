package ssh

import (
	"net"

	ice "github.com/shylinux/icebergs"
	"golang.org/x/crypto/ssh"
)

type Winsize struct{ Height, Width, x, y uint16 }

func _ssh_size(fd uintptr, b []byte) {
}
func _ssh_handle(m *ice.Message, meta map[string]string, c net.Conn, channel ssh.Channel, requests <-chan *ssh.Request) {
}
