package xterm

import (
	"os"
	"os/exec"

	pty "shylinux.com/x/creackpty"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type XTerm struct {
	*exec.Cmd
	*os.File
}

func (s XTerm) Setsize(rows, cols string) error {
	return pty.Setsize(s.File, &pty.Winsize{Rows: uint16(kit.Int(rows)), Cols: uint16(kit.Int(cols))})
}
func (s XTerm) Writeln(data string, arg ...ice.Any) {
	s.Write(kit.Format(data, arg...) + ice.NL)
}
func (s XTerm) Write(data string) (int, error) {
	return s.File.Write([]byte(data))
}
func (s XTerm) Close() error {
	return s.Cmd.Process.Kill()
}
func Command(m *ice.Message, dir string, cli string, arg ...string) (XTerm, error) {
	cmd := exec.Command(cli, arg...)
	cmd.Dir = nfs.MkdirAll(m, kit.Path(dir))
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "TERM=xterm")
	tty, err := pty.Start(cmd)
	return XTerm{cmd, tty}, err
}
