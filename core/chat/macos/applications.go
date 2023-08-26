package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	"shylinux.com/x/icebergs/core/team"
	kit "shylinux.com/x/toolkits"
)

const APPLICATIONS = "applications"

func init() {
	Index.MergeCommands(ice.Commands{
		APPLICATIONS: {Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				Notify(m, cli.RUNTIME, "系统启动成功", ctx.INDEX, cli.RUNTIME)
				FinderAppend(m, "Applications", m.PrefixKey())
				AppInstall(m, "Finder", nfs.DIR)
				AppInstall(m, "Safari", web.CHAT_IFRAME)
				AppInstall(m, "Terminal", web.CODE_XTERM)
				AppInstall(m, "Calendar", web.TEAM_PLAN, ctx.ARGS, team.MONTH)
				AppInstall(m, "Grapher", web.WIKI_DRAW)
				AppInstall(m, "Photos", web.WIKI_FEEL)
				AppInstall(m, "Books", web.WIKI_WORD)
				AppInstall(m, "", cli.RUNTIME, mdb.ICON, "usr/icons/info.png")
				AppInstall(m, "", web.DREAM, mdb.ICON, "usr/icons/Mission Control.png")
				AppInstall(m, "", web.CODE_VIMER, mdb.ICON, "usr/icons/vimer.png")
				AppInstall(m, "", web.CHAT_FLOWS, mdb.ICON, "usr/icons/flows.png")
				AppInstall(m, "", web.CODE_COMPILE, mdb.ICON, "usr/icons/go.png")
				AppInstall(m, "", web.CODE_GIT_STATUS, mdb.ICON, "usr/icons/git.png")
			}},
			code.INSTALL: {Hand: func(m *ice.Message, arg ...string) { AppInstall(m, arg[0], arg[1], arg[2:]...) }},
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
func install(m *ice.Message, cmd, name, index string, arg ...string) {
	name, icon := kit.Select(kit.Select("", kit.Split(index, ice.PT), -1), name), ""
	kit.If(nfs.Exists(m, kit.PathJoin(USR_ICONS, name, nfs.PNG)), func() { icon = kit.PathJoin(USR_ICONS, name, nfs.PNG) })
	m.Cmd(Prefix(cmd), mdb.CREATE, mdb.NAME, name, mdb.ICON, icon, ctx.INDEX, index, arg)
}
func AppInstall(m *ice.Message, name, index string, arg ...string) {
	install(m, APPLICATIONS, name, index, arg...)
}
