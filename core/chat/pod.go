package chat

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const POD = "pod"

func init() {
	Index.MergeCommands(ice.Commands{
		POD: {Help: "空间", Actions: web.ApiWhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			if m.IsCliUA() {
				if len(arg) == 0 || arg[0] == "" {
					m.Option(ice.MSG_USERROLE, aaa.TECH)
					list := m.CmdMap(web.DREAM, mdb.NAME)
					m.Cmd(web.SPACE, func(value ice.Maps) {
						msg := m.Cmd(nfs.DIR, path.Join(ice.USR_LOCAL_WORK, value[mdb.NAME], ice.USR_PUBLISH, kit.Keys(ice.ICE, m.OptionDefault(cli.GOOS, cli.LINUX), m.OptionDefault(cli.GOARCH, cli.AMD64))))
						kit.If(msg.Length() > 0, func() {
							m.Push(mdb.ICONS, list[value[mdb.NAME]][mdb.ICONS])
							m.Push(mdb.NAME, value[mdb.NAME]).Copy(msg)
						})
					})
					m.Cut("icons,name,size,time")
					m.RenderResult()
				} else if len(arg) > 1 {
					m.Option(ice.MSG_USERPOD, arg[0])
					m.Cmdy(web.SPACE, arg[0], arg[2], arg[3:])
				} else if strings.HasPrefix(m.Option(ice.MSG_USERUA), "git/") {
					m.RenderRedirect(kit.MergeURL2(m.Cmdv(web.SPACE, arg[0], web.CODE_GIT_REPOS, nfs.REMOTE, nfs.REMOTE), "/info/refs", m.OptionSimple("service")))
				} else if m.Option(cli.GOOS) != "" && m.Option(cli.GOARCH) != "" {
					m.RenderDownload(path.Join(ice.USR_LOCAL_WORK, arg[0], ice.USR_PUBLISH, kit.Keys(ice.ICE, m.Option(cli.GOOS), m.Option(cli.GOARCH))))
				} else {
					m.RenderDownload(path.Join(ice.USR_LOCAL_WORK, arg[0], ice.BIN_ICE_BIN))
				}
			} else if len(arg) == 0 || arg[0] == "" {
				web.RenderMain(m)
			} else {
				msg := m.Cmd(web.SPACE, arg[0])
				if msg.Length() == 0 && nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, arg[0])) {
					m.Cmd(web.DREAM, cli.START, kit.Dict(mdb.NAME, arg[0]))
				}
				if m.Option(ice.MSG_USERPOD, arg[0]); len(arg) == 1 {
					m.Cmdy(web.SPACE, arg[0], web.SPACE, ice.MAIN)
				} else if kit.IsIn(arg[1], CMD, "c") {
					if m.R.Method == "POST" || kit.IsIn(arg[2], web.ADMIN) {
						m.Cmdy(web.SPACE, arg[0], arg[2], arg[3:])
					} else {
						m.Options(msg.AppendSimple()).Options(mdb.ICONS, "")
						web.RenderPodCmd(m, arg[0], arg[2], arg[3:])
					}
				}
			}
		}},
	})
}
