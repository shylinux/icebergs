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
				mdb.IsSearchPreview(m, arg, func() []string { return []string{web.LINK, PORTAL, m.MergePodCmd("", "")} })
			}},
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) { web.RenderCmd(m, "", arg) }},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if p := path.Join(ice.USR_PORTAL, path.Join(arg...)); (m.Option(ice.DEBUG) == ice.TRUE || !nfs.ExistsFile(m, p)) && aaa.Right(m.Spawn(), arg) {
					ctx.Run(m, arg...)
					m.Cmd(nfs.SAVE, p, ice.Maps{nfs.CONTENT: m.FormatMeta(), nfs.DIR_ROOT: ""})
				} else {
					m.Copy(m.Spawn([]byte(m.Cmdx(nfs.CAT, p))))
				}
			}},
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.Switch(m.Option(mdb.TYPE), kit.Simple(web.SERVER, web.WORKER), func() {
					m.PushButton(ice.Maps{PORTAL: "官网"})
				})
			}},
		}, aaa.WhiteAction(ctx.COMMAND, ctx.RUN), aaa.RoleAction(ctx.COMMAND, ctx.RUN), web.DreamAction(), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Push(HEADER, m.Cmdx(WORD, path.Join(nfs.SRC_DOCUMENT, INDEX_SHY))); len(arg) > 0 {
				m.Push(NAV, m.Cmdx(WORD, path.Join(nfs.SRC_DOCUMENT, path.Join(arg...), INDEX_SHY)))
			}
			m.Display("")
		}},
	})
}
