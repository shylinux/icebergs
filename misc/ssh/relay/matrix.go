package relay

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/web"
)

type matrix struct {
	list string `data:"list refresh"`
}

func (s matrix) List(m *ice.Message, arg ...string) *ice.Message {
	m.Cmdy("ssh.relay", "dream").Display("")
	m.Sort("type,status,space,machine", []string{web.SERVER, web.WORKER, ""}, []string{cli.START, cli.STOP, ""}, "str_r", "str_r")
	return m
}
func init() { ice.Cmd("ssh.matrix", matrix{}) }
