package chat

import (
	"path"
	"strings"

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
		POD: {Actions: ice.MergeActions(ctx.CmdAction(), web.ApiAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || kit.Select("", arg, 0) == "" {
				web.RenderCmd(m, web.SPACE)
			} else if strings.HasPrefix(m.Option(ice.MSG_USERUA), "git/") {
				m.RenderRedirect(m.Cmdv(web.SPACE, arg[0], "web.code.git.repos", nfs.REMOTE, nfs.REMOTE) + "/info/refs?service=" + m.Option("service"))
			} else if m.Option(cli.GOOS) != "" && m.Option(cli.GOARCH) != "" {
				m.RenderDownload(path.Join(ice.USR_LOCAL_WORK, arg[0], ice.USR_PUBLISH, kit.Keys(ice.ICE, m.Option(cli.GOOS), m.Option(cli.GOARCH))))
			} else if m.IsCliUA() {
				m.RenderDownload(path.Join(ice.USR_LOCAL_WORK, arg[0], ice.BIN_ICE_BIN))
			} else {
				if m.Cmd(web.SPACE, arg[0]).Length() == 0 && nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, arg[0])) {
					m.Cmd(web.DREAM, cli.START, kit.Dict(mdb.NAME, arg[0]))
				}
				if len(arg) == 1 {
					m.Cmdy(web.SPACE, arg[0], web.SPACE, ice.MAIN, kit.Dict(nfs.VERSION, web.RenderVersion(m)))
				} else if arg[1] == CMD {
					web.RenderPodCmd(m, arg[0], arg[2], arg[3:])
				}
			}
		}},
	})
}
