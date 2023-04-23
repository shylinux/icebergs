//go:build windows
// +build windows

package xterm

import (
	"errors"
	"os"
	"os/exec"
)

func Setsid(cmd *exec.Cmd)                 {}
func Setsize(*os.File, *Winsize) error     { return errors.New("unsupported") }
func Open() (pty, tty *os.File, err error) { return nil, nil, errors.New("unsupported") }
