package ssh

import (
	"io"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/misc/xterm"
	kit "shylinux.com/x/toolkits"
)

type XTerm struct {
	io.Reader
	io.Writer
	*ssh.Session
}

func (s XTerm) Setsize(rows, cols string) error {
	return s.Session.WindowChange(kit.Int(rows), kit.Int(cols))
}
func (s XTerm) Close() error {
	return s.Session.Close()
}

func Shell(m *ice.Message, cb func(xterm.XTerm)) {
	_ssh_conn(m, func(c *ssh.Client) {
		if s, e := c.NewSession(); !m.Warn(e, ice.ErrNotValid) {
			defer s.Close()
			w, _ := s.StdinPipe()
			r, _ := s.StdoutPipe()
			width, height, _ := terminal.GetSize(int(os.Stdin.Fd()))
			s.RequestPty(kit.Env(cli.TERM), height, width, ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400})
			defer s.Wait()
			s.Shell()
			cb(&XTerm{r, w, s})
		}
	})
}
