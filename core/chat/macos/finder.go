package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const FINDER = "finder"

func init() {
	Index.MergeCommands(ice.Commands{
		FINDER: {Name: "finder list", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				mdb.IsSearchPreview(m, arg, func() []string { return []string{web.LINK, DESKTOP, m.MergePodCmd("", DESKTOP)} })
			}},
		}, CmdHashAction(mdb.NAME))},
	})
}

func FinderAppend(m *ice.Message, name, index string, arg ...string) {
	install(m, FINDER, name, index, arg...)
}
