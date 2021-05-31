package ssh

import (
	"encoding/binary"
	"github.com/kr/pty"
	"io"
	"net"
	"os"
	"syscall"
	"unsafe"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
	"golang.org/x/crypto/ssh"
)

func _ssh_size(fd uintptr, b []byte) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])

	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}
func _ssh_handle(m *ice.Message, meta map[string]string, c net.Conn, channel ssh.Channel, requests <-chan *ssh.Request) {
	m.Logs(CHANNEL, tcp.HOSTPORT, c.RemoteAddr(), "->", c.LocalAddr())
	defer m.Logs("dischan", tcp.HOSTPORT, c.RemoteAddr(), "->", c.LocalAddr())

	shell := kit.Select("bash", os.Getenv("SHELL"))
	list := []string{cli.PATH + "=" + os.Getenv(cli.PATH)}

	pty, tty, err := pty.Open()
	if m.Warn(err != nil, err) {
		return
	}
	defer tty.Close()

	h := m.Rich(CHANNEL, "", kit.Data(kit.MDB_STATUS, tcp.OPEN, TTY, tty.Name(), meta))
	meta[CHANNEL] = h

	for request := range requests {
		m.Logs(REQUEST, tcp.HOSTPORT, c.RemoteAddr(), kit.MDB_TYPE, request.Type)

		switch request.Type {
		case "pty-req":
			termLen := request.Payload[3]
			termEnv := string(request.Payload[4 : termLen+4])
			_ssh_size(pty.Fd(), request.Payload[termLen+4:])
			list = append(list, "TERM="+termEnv)

		case "window-change":
			_ssh_size(pty.Fd(), request.Payload)

		case "env":
			var env struct{ Name, Value string }
			if err := ssh.Unmarshal(request.Payload, &env); err != nil {
				continue
			}
			list = append(list, env.Name+"="+env.Value)

		case "exec":
			_ssh_exec(m, shell, []string{"-c", string(request.Payload[4 : request.Payload[3]+4])}, list, channel, func() {
				channel.Close()
			})
		case "shell":
			m.Go(func() { io.Copy(channel, pty) })

			_ssh_exec(m, shell, nil, list, tty, func() {
				defer m.Cmd(mdb.MODIFY, CHANNEL, "", mdb.HASH, kit.MDB_HASH, h, kit.MDB_STATUS, tcp.CLOSE)
				_ssh_close(m, c, channel)
			})

			_ssh_watch(m, meta, h, channel, pty, channel)
		}
		request.Reply(true, nil)
	}
}
