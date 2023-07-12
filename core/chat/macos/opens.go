package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

const OPENS = "opens"

func init() {
	Index.MergeCommands(ice.Commands{
		OPENS: {Name: "opens app auto", Hand: func(m *ice.Message, arg ...string) {
			if tcp.IsLocalHost(m, m.Option(ice.MSG_USERIP)) {
				arg[0] = kit.ExtChange(arg[0], "app")
				cli.Opens(m, arg...)
			}
		}},
	})
}
