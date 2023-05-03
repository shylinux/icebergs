package macos

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
)

const OPENS = "opens"

func init() {
	Index.MergeCommands(ice.Commands{
		OPENS: {Name: "open app auto", Hand: func(m *ice.Message, arg ...string) {
			if strings.HasPrefix(m.Option(ice.MSG_USERWEB), "http://localhost:") {
				m.Cmd(cli.SYSTEM, "open", "-a", arg)
			}
		}},
	})
}
