package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const DESKTOP = "desktop"

func init() {
	Index.MergeCommands(ice.Commands{
		DESKTOP: {Help: "应用桌面", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				if mdb.HashImport(m); m.Cmd(DESKTOP).Length() == 0 {
					DeskAppend(m, "usr/icons/Books.png", web.WIKI_WORD)
					DeskAppend(m, "usr/icons/Photos.png", web.WIKI_FEEL)
					DeskAppend(m, "usr/icons/Grapher.png", web.WIKI_DRAW)
					DeskAppend(m, "usr/icons/Calendar.png", web.TEAM_PLAN)
				}
			}},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { mdb.HashExport(m) }},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), kit.Simple(web.SERVER, web.WORKER), func() { m.PushButton(kit.Dict(m.CommandKey(), "桌面")) })
			}},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcess(m, []string{}, arg...) }},
		}, aaa.RoleAction(), PodCmdAction(), CmdHashAction(), mdb.ImportantHashAction())},
	})
}

func DeskAppend(m *ice.Message, icon, index string) {
	install(m, DESKTOP, icon, index)
}
