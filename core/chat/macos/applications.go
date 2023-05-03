package macos

import (
	ice "shylinux.com/x/icebergs"
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
				Notify(m, "runtime", "系统启动成功", ctx.INDEX, "cli.runtime")
				FinderAppend(m, "Applications", m.PrefixKey())
				FinderAppend(m, "Pictures", web.WIKI_FEEL)
				AppInstall(m, "Finder", "nfs.dir")
				AppInstall(m, "Safari", web.CHAT_IFRAME)
				AppInstall(m, "Calendar", web.TEAM_PLAN, ctx.ARGS, team.MONTH)
				AppInstall(m, "Terminal", web.CODE_XTERM)
				AppInstall(m, "Grapher", web.WIKI_DRAW)
				AppInstall(m, "Photos", web.WIKI_FEEL)
				AppInstall(m, "Books", web.WIKI_WORD)
			}},
			code.INSTALL: {Hand: func(m *ice.Message, arg ...string) { AppInstall(m, arg[0], arg[1], arg[2:]...) }},
		}, CmdHashAction("index,args"))},
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
