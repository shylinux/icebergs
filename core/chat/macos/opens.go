package macos

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
)

const OPENS = "opens"

func init() {
	Index.MergeCommands(ice.Commands{
		OPENS: {Name: "opens app auto", Hand: func(m *ice.Message, arg ...string) {
			if strings.HasPrefix(m.Option(ice.MSG_USERWEB), "http://localhost:") {
				cli.Opens(m, arg...)
			}
		}},
	})
}
