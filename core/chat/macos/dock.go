package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const DOCK = "dock"

func init() {
	Index.MergeCommands(ice.Commands{DOCK: {Actions: ice.MergeActions(ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashImport(m); m.Cmd(DOCK).Length() == 0 {
				DockAppend(m, "usr/icons/Finder.png", Prefix(FINDER))
				DockAppend(m, "usr/icons/Safari.png", web.CHAT_IFRAME)
				DockAppend(m, "usr/icons/Terminal.png", web.CODE_XTERM)
				DockAppend(m, "usr/icons/vimer.png", web.CODE_VIMER)
				DockAppend(m, "usr/icons/go.png", web.CODE_COMPILE)
				DockAppend(m, "usr/icons/git.png", web.CODE_GIT_STATUS)
			}
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, arg ...string) { mdb.HashExport(m) }},
	}, aaa.RoleAction(), CmdHashAction(), mdb.ImportantHashAction())}})
}

func DockAppend(m *ice.Message, icon, index string) {
	install(m, DOCK, icon, index)
}
