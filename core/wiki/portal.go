package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

const PORTAL = "portal"

func init() {
	const (
		HEADER    = "header"
		NAV       = "nav"
		INDEX_SHY = "index.shy"
	)
	Index.MergeCommands(ice.Commands{
		PORTAL: {Name: "portal path auto", Help: "网站/门户", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				mdb.IsSearchForEach(m, arg, func() []string { return []string{web.LINK, PORTAL, m.MergePodCmd("", "") + nfs.PS} })
			}},
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) { web.RenderCmd(m, "", arg) }},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Push(HEADER, m.Cmdx(WORD, path.Join(ice.SRC_DOCUMENT, INDEX_SHY))); len(arg) > 0 {
				m.Push(NAV, m.Cmdx(WORD, path.Join(ice.SRC_DOCUMENT, path.Join(arg...), INDEX_SHY)))
			}
			m.Display("")
		}},
	})
}
