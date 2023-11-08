package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const DOCK = "dock"

func init() {
	Index.MergeCommands(ice.Commands{
		DOCK: {Help: "工具栏", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd(DOCK).Length() == 0 {
					DockAppend(m, "Finder.png", Prefix(FINDER))
					DockAppend(m, "Safari.png", web.CHAT_IFRAME)
					DockAppend(m, "Terminal.png", web.CODE_XTERM)
					DockAppend(m, "go.png", web.CODE_COMPILE)
					DockAppend(m, "git.png", web.CODE_GIT_STATUS)
					DockAppend(m, "vimer.png", web.CODE_VIMER)
				}
			}},
		}, aaa.RoleAction(), CmdHashAction(), mdb.ExportHashAction())},
	})
}

func DockAppend(m *ice.Message, icon, index string) { install(m, DOCK, icon, index) }
