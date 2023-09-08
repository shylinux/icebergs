package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _feel_path(m *ice.Message, p string) string {
	if nfs.Exists(m, ice.USR_LOCAL_IMAGE) {
		return path.Join(ice.USR_LOCAL_IMAGE, p)
	}
	return p
}

const FEEL = "feel"

func init() {
	Index.MergeCommands(ice.Commands{
		FEEL: {Name: "feel path auto prev next record1 record2 upload actions", Icon: "usr/icons/Photos.png", Help: "影音媒体", Actions: ice.MergeActions(ice.Actions{
			"record1": {Help: "截图"}, "record2": {Help: "录屏"},
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.PATH, _feel_path(m, m.Option(nfs.PATH)))
				up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
				m.Cmdy(web.CACHE, web.WATCH, m.Option(mdb.HASH), path.Join(m.Option(nfs.PATH), up[1]))
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Trash(m, _feel_path(m, m.Option(nfs.PATH)))
			}},
		}, WikiAction("", "png|PNG|jpg|JPG|jpeg|mp4|m4v|mov|MOV|webm")), Hand: func(m *ice.Message, arg ...string) {
			m.Option(nfs.DIR_ROOT, _feel_path(m, ""))
			_wiki_list(m, kit.Slice(arg, 0, 1)...)
			ctx.DisplayLocal(m, "")
		}},
	})
}
