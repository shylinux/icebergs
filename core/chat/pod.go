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
		POD: {Name: "pod", Help: "节点", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, POD) }},
		}, ctx.CmdAction(), web.ApiAction()), Hand: func(m *ice.Message, arg ...string) {
			if web.OptionAgentIs(m, "curl", "Wget") {
				aaa.UserRoot(m)
				m.Option(ice.POD, kit.Select("", arg, 0))
				m.Cmdy(web.SHARE_LOCAL, ice.BIN_ICE_BIN)
				return // 下载程序
			}

			if len(arg) == 0 || kit.Select("", arg, 0) == "" {
				web.RenderCmd(m, web.ROUTE) // 节点列表

			} else if len(arg) == 1 {
				if m.Cmd(web.SPACE, arg[0]).Length() == 0 && !strings.Contains(arg[0], ice.PT) {
					m.Cmd(web.DREAM, cli.START, mdb.NAME, arg[0]) // 启动节点
				}
				aaa.UserRoot(m)
				if web.RenderWebsite(m, arg[0], ice.INDEX_IML, "Header", "", "River", "", "Footer", ""); m.Result() == "" {
					web.RenderIndex(m, ice.VOLCANOS) // 节点首页
				}

			} else if arg[1] == CMD {
				m.Cmdy(web.SPACE, arg[0], m.Prefix(CMD), path.Join(arg[2:]...)) // 节点命令

			} else if arg[1] == WEBSITE {
				web.RenderWebsite(m, arg[0], path.Join(arg[2:]...)) // 节点网页
			}
		}},
	})
}
