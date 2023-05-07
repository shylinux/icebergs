package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const DOCK = "dock"

func init() {
	Index.MergeCommands(ice.Commands{DOCK: {Actions: ice.MergeActions(ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			if m.Cmd(DOCK).Length() == 0 {
				DockAppend(m, "Finder", Prefix(FINDER))
				DockAppend(m, "Safari", web.CHAT_IFRAME)
				DockAppend(m, "Terminal", web.CODE_XTERM)
				DockAppend(m, "", web.CODE_GIT_REPOS, mdb.ICON, "usr/icons/git.jpg")
				DockAppend(m, "", web.CODE_COMPILE, mdb.ICON, "usr/icons/go.png")
				DockAppend(m, "", web.CODE_VIMER)
				DockAppend(m, "", web.DREAM)
			}
		}},
	}, mdb.ImportantHashAction(), CmdHashAction())}})
}

func DockAppend(m *ice.Message, name, index string, arg ...string) {
	install(m, DOCK, name, index, arg...)
}
