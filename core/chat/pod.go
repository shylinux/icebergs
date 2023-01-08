package chat

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const POD = "pod"

func init() {
	Index.MergeCommands(ice.Commands{
		POD: {Name: "pod", Help: "节点", Actions: ice.MergeActions(ice.Actions{
			web.SERVE_PARSE: {Hand: func(m *ice.Message, arg ...string) {
				if kit.Select("", arg, 0) == CHAT {
					kit.Fetch(arg[1:], func(k, v string) { m.Logs("Refer", k, v).Options(k, v) })
				}
			}},
		}, ctx.CmdAction(), web.ServeAction(), web.ApiAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if web.OptionAgentIs(m, "curl", "wget") {
				m.Cmdy(web.SHARE_LOCAL, ice.BIN_ICE_BIN, kit.Dict(ice.POD, kit.Select("", arg, 0), ice.MSG_USERROLE, aaa.TECH))
				return
			}
			if len(arg) == 0 || kit.Select("", arg, 0) == "" {
				web.RenderCmd(m, web.SPACE)
			} else if len(arg) == 1 {
				if m.Cmd(web.SPACE, arg[0]).Length() == 0 && nfs.ExistsFile(m, path.Join(ice.USR_LOCAL_WORK, arg[0])) {
					m.Cmd(web.DREAM, cli.START, kit.Dict(mdb.NAME, arg[0]))
				}
				web.RenderMain(aaa.UserRoot(m), arg[0], "")
			} else if arg[1] == CMD {
				web.RenderPodCmd(m, arg[0], arg[2], arg[3:])
			} else if arg[1] == WEBSITE {
				RenderWebsite(m, arg[0], path.Join(arg[2:]...))
			}
		}},
	})
}
