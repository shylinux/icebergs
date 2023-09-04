package macos

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const APPLICATIONS = "applications"

func init() {
	Index.MergeCommands(ice.Commands{
		APPLICATIONS: {Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				Notify(m, cli.RUNTIME, "系统启动成功", ctx.INDEX, cli.RUNTIME)
				FinderAppend(m, "Applications", m.PrefixKey())
				AppInstall(m, "usr/icons/dir.png", nfs.DIR)
				AppInstall(m, "usr/icons/Safari.png", web.CHAT_IFRAME)
				AppInstall(m, "usr/icons/Terminal.png", web.CODE_XTERM)
				AppInstall(m, "usr/icons/Calendar.png", web.TEAM_PLAN)
				AppInstall(m, "usr/icons/Grapher.png", web.WIKI_DRAW)
				AppInstall(m, "usr/icons/Photos.png", web.WIKI_FEEL)
				AppInstall(m, "usr/icons/Books.png", web.WIKI_WORD)

				AppInstall(m, "usr/icons/info.png", cli.RUNTIME)
				AppInstall(m, "usr/icons/Mission Control.png", web.DREAM, mdb.NAME, web.DREAM)
				AppInstall(m, "usr/icons/vimer.png", web.CODE_VIMER)
				AppInstall(m, "usr/icons/flows.png", web.CHAT_FLOWS)
				AppInstall(m, "usr/icons/go.png", web.CODE_COMPILE)
				AppInstall(m, "usr/icons/git.png", web.CODE_GIT_STATUS)
			}},
			code.INSTALL: {Hand: func(m *ice.Message, arg ...string) { AppInstall(m, arg[0], arg[1]) }},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case web.SPACE:
					m.Cmdy(web.SPACE).CutTo(mdb.NAME, arg[0])
				case ctx.INDEX:
					m.Cmdy(web.Space(m, m.Option(web.SPACE)), ctx.COMMAND)
				case ctx.ARGS:
					m.Cmdy(web.Space(m, m.Option(web.SPACE)), ctx.COMMAND, mdb.INPUTS, m.Option(ctx.INDEX))
				case mdb.ICON:
					if m.Option(ctx.INDEX) != "" {
						m.Cmd(web.Space(m, m.Option(web.SPACE)), m.Option(ctx.INDEX), mdb.INPUTS, arg[0]).Table(func(value ice.Maps) {
							m.Push(arg[0], value[arg[0]]+"?pod="+kit.Keys(m.Option(ice.MSG_USERPOD), m.Option(web.SPACE)))
						})
					}
					m.Cmd(nfs.DIR, USR_ICONS, func(value ice.Maps) { m.Push(arg[0], value[nfs.PATH]) })
				}
			}}, mdb.CREATE: {Name: "create space index args name icon"},
		}, PodCmdAction(), CmdHashAction("space,index,args"))},
	})
}
func install(m *ice.Message, cmd, icon, index string, arg ...string) {
	name := kit.TrimExt(path.Base(icon), "png")
	m.Cmd(Prefix(cmd), mdb.CREATE, mdb.NAME, name, mdb.ICON, icon, ctx.INDEX, index, arg)
}
func AppInstall(m *ice.Message, icon, index string, arg ...string) {
	install(m, APPLICATIONS, icon, index, arg...)
}
