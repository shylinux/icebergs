//go:build windows
// +build windows

package pty

import (
	"errors"
	"os"
)

func Setsize(*os.File, *Winsize) error     { return errors.New("unsupported") }
func Open() (pty, tty *os.File, err error) { return nil, nil, errors.New("unsuported") }
