package chat

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const POD = "pod"

func init() {
	Index.MergeCommands(ice.Commands{
		web.PP(POD): {Name: "/pod/", Help: "节点", Actions: ice.MergeActions(ctx.CmdAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if web.OptionAgentIs(m, "curl", "wget") {
				aaa.UserRoot(m).Cmdy(web.SHARE_LOCAL, ice.BIN_ICE_BIN, kit.Dict(ice.POD, kit.Select("", arg, 0)))
				return
			}
			if len(arg) == 0 || kit.Select("", arg, 0) == "" {
				web.RenderCmd(m, web.ROUTE)
			} else if len(arg) == 1 {
				if m.Cmd(web.SPACE, arg[0]).Length() == 0 && !strings.Contains(arg[0], ice.PT) {
					m.Cmd(web.DREAM, cli.START, mdb.NAME, arg[0])
				}
				aaa.UserRoot(m)
				if web.RenderWebsite(m, arg[0], ice.INDEX_IML, "Header", "", "River", "", "Footer", ""); m.Result() == "" {
					web.RenderIndex(m, ice.VOLCANOS)
				}
			} else if arg[1] == CMD {
				m.Cmdy(web.SPACE, arg[0], m.Prefix(CMD), path.Join(arg[2:]...))
			} else if arg[1] == WEBSITE {
				web.RenderWebsite(m, arg[0], path.Join(arg[2:]...))
			}
		}},
	})
}
