package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const FINDER = "finder"

func init() {
	Index.MergeCommands(ice.Commands{
		FINDER: {Name: "finder list", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd(DOCK).Length() == 0 {
					DockAppend(m, "Finder", m.PrefixKey())
					DockAppend(m, "Safari", web.CHAT_IFRAME)
					DockAppend(m, "Terminal", web.CODE_XTERM)
					DockAppend(m, "", web.CODE_GIT_STATUS, mdb.ICON, "usr/icons/git.jpg")
					DockAppend(m, "", web.CODE_COMPILE, mdb.ICON, "usr/icons/go.png")
					DockAppend(m, "", web.CODE_VIMER)
				}
			}},
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				mdb.IsSearchForEach(m, arg, func() []string {
					return []string{web.LINK, DESKTOP, m.MergePodCmd("", DESKTOP, log.DEBUG, ice.TRUE)}
				})
			}},
		}, CmdHashAction(mdb.NAME))},
	})
}

func FinderAppend(m *ice.Message, name, index string, arg ...string) {
	install(m, FINDER, name, index, arg...)
}
