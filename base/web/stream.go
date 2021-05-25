package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
)

type Buffer struct {
	m *ice.Message
	n string
}

func (b *Buffer) Write(buf []byte) (int, error) {
	b.m.Cmd(SPACE, b.n, "grow", string(buf))
	return len(buf), nil
}
func (b *Buffer) Close() error { return nil }

func PushStream(m *ice.Message) {
	m.Option(cli.CMD_OUTPUT, &Buffer{m: m, n: m.Option(ice.MSG_DAEMON)})
}
