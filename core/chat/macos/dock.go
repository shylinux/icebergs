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
				DockAppend(m, "", web.CODE_VIMER)
			}
		}},
	}, CmdHashAction(), mdb.ImportantHashAction())}})
}

func DockAppend(m *ice.Message, name, index string, arg ...string) {
	install(m, DOCK, name, index, arg...)
}
