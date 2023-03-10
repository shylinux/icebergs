package chat

import (
	"runtime"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const SEARCH = "search"

func init() {
	Index.MergeCommands(ice.Commands{
		web.P(SEARCH): {Name: "/search", Help: "搜索框", Actions: ice.MergeActions(ice.Actions{
			cli.OPENS: {Hand: func(m *ice.Message, arg ...string) {
				switch runtime.GOOS {
				case cli.DARWIN:
					if kit.Ext(arg[0]) == "app" {
						m.Cmd(cli.SYSTEM, cli.OPEN, "-a", arg[0])
					} else {
						m.Cmd(cli.SYSTEM, cli.OPEN, arg[0])
					}
				case cli.WINDOWS:
					if kit.Ext(arg[0]) == "exe" {
						m.Cmd(cli.SYSTEM, arg[0])
					} else {
						m.Cmd(cli.SYSTEM, "explorer", arg[0])
					}
				}
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(web.Space(m, m.Option(ice.POD)), mdb.SEARCH, arg).StatusTimeCount()
		}},
	})
}
