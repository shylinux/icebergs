package xterm

import (
	"os"
	"os/exec"
	"syscall"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type Winsize struct {
	Rows uint16 // ws_row: Number of rows (in cells)
	Cols uint16 // ws_col: Number of columns (in cells)
	X    uint16 // ws_xpixel: Width in pixels
	Y    uint16 // ws_ypixel: Height in pixels
}

type XTerm struct {
	*exec.Cmd
	*os.File
}

func (s XTerm) Setsize(rows, cols string) error {
	return Setsize(s.File, &Winsize{Rows: uint16(kit.Int(rows)), Cols: uint16(kit.Int(cols))})
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
func Command(m *ice.Message, dir string, cli string, arg ...string) (*XTerm, error) {
	cmd := exec.Command(cli, arg...)
	cmd.Dir = nfs.MkdirAll(m, kit.Path(dir))
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "TERM=xterm")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}
	pty, tty, err := Open()
	if err != nil {
		return nil, err
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, tty, tty
	return &XTerm{cmd, pty}, cmd.Start()
}
