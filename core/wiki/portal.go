package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

const (
	INDEX_SHY = "index.shy"
	HEADER    = "header"
	NAV       = "nav"
)

const PORTAL = "portal"

func init() {
	Index.MergeCommands(ice.Commands{
		PORTAL: {Name: "portal path auto", Help: "官网", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { aaa.White(m, ice.VAR_PORTAL) }},
			ctx.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if p := path.Join(ice.VAR_PORTAL, path.Join(arg...)); (m.Option(ice.DEBUG) == ice.TRUE || !nfs.ExistsFile(m, p)) && aaa.Right(m.Spawn(), arg) {
					ctx.Run(m, arg...)
					m.Cmd(nfs.SAVE, p, ice.Maps{nfs.CONTENT: m.FormatsMeta(nil), nfs.DIR_ROOT: ""})
				} else {
					// m.Option(ice.MSG_USERROLE, aaa.TECH)
					m.Copy(m.Spawn([]byte(m.Cmdx(nfs.CAT, p))))
				}
			}},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) { web.DreamProcessIframe(m, arg...) }},
		}, web.ServeCmdAction(), web.DreamTablesAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Push(HEADER, m.Cmdx(WORD, _portal_path(m, INDEX_SHY))); len(arg) > 0 {
				m.Push(NAV, m.Cmdx(WORD, _portal_path(m, path.Join(arg...), INDEX_SHY)))
			}
			m.Display("")
		}},
	})
}
func _portal_path(m *ice.Message, arg ...string) (res string) {
	if !nfs.Exists(m, path.Join(nfs.SRC_DOCUMENT, path.Join(arg...)), func(p string) { res = p }) {
		res = path.Join(nfs.USR_LEARNING_PORTAL, path.Join(arg...))
	}
	return res
}
