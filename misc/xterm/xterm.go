package xterm

import (
	"os"
	"os/exec"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type Winsize struct {
	Rows uint16 // ws_row: Number of rows (in cells)
	Cols uint16 // ws_col: Number of columns (in cells)
	X    uint16 // ws_xpixel: Width in pixels
	Y    uint16 // ws_ypixel: Height in pixels
}

type XTerm interface {
	Setsize(rows, cols string) error
	Writeln(data string, arg ...ice.Any)
	Write(data string) (int, error)
	Read(buf []byte) (int, error)
	Close() error
}
type xterm struct {
	*exec.Cmd
	*os.File
}

func (s xterm) Setsize(rows, cols string) error {
	return Setsize(s.File, &Winsize{Rows: uint16(kit.Int(rows)), Cols: uint16(kit.Int(cols))})
}
func (s xterm) Writeln(data string, arg ...ice.Any) { s.Write(kit.Format(data, arg...) + lex.NL) }
func (s xterm) Write(data string) (int, error)      { return s.File.Write([]byte(data)) }
func (s xterm) Read(buf []byte) (int, error)        { return s.File.Read(buf) }
func (s xterm) Close() error                        { return s.Cmd.Process.Kill() }

func Command(m *ice.Message, dir string, cli string, arg ...string) (XTerm, error) {
	if path.Base(cli) == "ish" {
		return newiterm(m)
	}
	cmd := exec.Command(cli, arg...)
	cmd.Dir = nfs.MkdirAll(m, kit.Path(dir))
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "TERM=xterm")
	if pty, tty, err := Open(); err != nil {
		return nil, err
	} else {
		Setsid(cmd)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, tty, tty
		return &xterm{cmd, pty}, cmd.Start()
	}
}
