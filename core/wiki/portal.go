package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const PORTAL = "portal"

func init() {
	const (
		HEADER    = "header"
		NAV       = "nav"
		INDEX_SHY = "index.shy"
	)
	Index.MergeCommands(ice.Commands{
		PORTAL: {Name: "portal path auto", Help: "网站门户", Actions: ice.MergeActions(ice.Actions{
			mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
				mdb.IsSearchPreview(m, arg, func() []string { return []string{web.LINK, PORTAL, m.MergePodCmd("", "") + nfs.PS} })
			}},
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) { web.RenderCmd(m, "", arg) }},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if p := path.Join(ice.USR_PORTAL, path.Join(arg...)); nfs.ExistsFile(m, p) && !(m.Option(ice.DEBUG) == ice.TRUE && aaa.Right(m.Spawn(), arg)) {
					m.Copy(m.Spawn([]byte(m.Cmdx(nfs.CAT, p))))
				} else {
					ctx.Run(m, arg...)
					m.Cmd(nfs.SAVE, p, kit.Dict(nfs.CONTENT, m.FormatMeta()))
				}
			}},
		}, aaa.RoleAction(ctx.COMMAND, ctx.RUN), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Push(HEADER, m.Cmdx(WORD, path.Join(nfs.SRC_DOCUMENT, INDEX_SHY))); len(arg) > 0 {
				m.Push(NAV, m.Cmdx(WORD, path.Join(nfs.SRC_DOCUMENT, path.Join(arg...), INDEX_SHY)))
			}
			m.Display("")
		}},
	})
}
