package xterm

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/task"
)

type Winsize struct {
	Rows uint16 // ws_row: Number of rows (in cells)
	Cols uint16 // ws_col: Number of columns (in cells)
	X    uint16 // ws_xpixel: Width in pixels
	Y    uint16 // ws_ypixel: Height in pixels
}

type XTerm interface {
	Setsize(rows, cols string) error
	Write(buf []byte) (int, error)
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
func (s xterm) Writeln(str string, arg ...ice.Any) { s.Write([]byte(kit.Format(str, arg...) + lex.NL)) }
func (s xterm) Write(buf []byte) (int, error)      { return s.File.Write(buf) }
func (s xterm) Read(buf []byte) (int, error)       { return s.File.Read(buf) }
func (s xterm) Close() error                       { s.Cmd.Process.Kill(); return s.File.Close() }

type handler func(m *ice.Message, arg ...string) (XTerm, error)

var list = map[string]handler{}

func AddCommand(key string, cb handler) { list[key] = cb }

func Command(m *ice.Message, dir string, cmd string, arg ...string) (XTerm, error) {
	if cb, ok := list[path.Base(cmd)]; ok {
		m.Debug("find shell %s %s", cmd, kit.FileLines(cb))
		return cb(m.Spawn(), arg...)
	}
	if m.Cmd(cli.SUDO, cmd).Length() > 0 {
		m.Debug("find sudo %s", cmd)
		cmd, arg = cli.SUDO, kit.Simple(cmd, arg)
	}
	p := exec.Command(cmd, arg...)
	p.Dir = nfs.MkdirAll(m, kit.Path(dir))
	p.Env = append(p.Env, os.Environ()...)
	p.Env = append(p.Env, "TERM=xterm")
	if pty, tty, err := Open(); err != nil {
		return nil, err
	} else {
		Setsid(p)
		p.Stdin, p.Stdout, p.Stderr = tty, tty, tty
		return &xterm{p, pty}, p.Start()
	}
}
func PushShell(m *ice.Message, xterm XTerm, cmds []string, cb func(string)) {
	list := [][]string{}
	list = append(list, []string{""})
	lock := task.Lock{}
	m.Go(func() {
		kit.For(cmds, func(cmd string) {
			for {
				m.Sleep300ms()
				if func() bool { defer lock.Lock()(); return len(list[len(list)-1]) > 1 }() {
					break
				}
			}
			m.Debug("cmd %v", cmd)
			fmt.Fprintln(xterm, cmd)
			defer lock.Lock()()
			list = append(list, []string{cmd})
		})
		m.Sleep(m.OptionDefault("interval", "3s"))
	})
	kit.For(xterm, func(res []byte) {
		m.Debug("res %v %v", string(res), res)
		cb(string(res))
		defer lock.Lock()()
		list[len(list)-1] = append(list[len(list)-1], string(res))
	})
}
