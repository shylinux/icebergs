package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/chat"
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
		FEEL: {Name: "feel path auto upload record1 record2 actions", Icon: "Photos.png", Help: "影音媒体", Actions: ice.MergeActions(ice.Actions{
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.PATH, _feel_path(m, m.Option(nfs.PATH)))
				up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
				m.Cmdy(web.CACHE, web.WATCH, m.Option(mdb.HASH), path.Join(m.Option(nfs.PATH), up[1]))
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Trash(m, kit.Select(_feel_path(m, m.Option(nfs.PATH)), arg, 0))
			}},
			chat.FAVOR_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.TYPE:
					m.Push(arg[0], web.IMAGE_PNG)
				case mdb.TEXT:
					kit.If(m.Option(mdb.TYPE) == web.IMAGE_PNG, func() {
						m.Cmdy(nfs.DIR, ice.USR_ICONS).CutTo(nfs.PATH, arg[0])
					})
				}
			}},
			chat.FAVOR_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				if html.IsImage(m.Option(mdb.NAME), m.Option(mdb.TYPE)) || html.IsVideo(m.Option(mdb.NAME), m.Option(mdb.TYPE)) || html.IsAudio(m.Option(mdb.NAME), m.Option(mdb.TYPE)) {
					m.PushButton(kit.Dict(m.CommandKey(), "预览"))
				}
			}},
			chat.FAVOR_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				if m.Option(ctx.ACTION) == m.CommandKey() {
					if link := web.SHARE_LOCAL + m.Option(mdb.TEXT); html.IsImage(m.Option(mdb.NAME), m.Option(mdb.TYPE)) {
						m.EchoImages(link)
					} else if html.IsVideo(m.Option(mdb.NAME), m.Option(mdb.TYPE)) {
						m.EchoVideos(link)
					} else if html.IsAudio(m.Option(mdb.NAME), m.Option(mdb.TYPE)) {
						m.EchoAudios(link)
					}
					m.ProcessInner()
				}
			}},
		}, chat.FavorAction(), WikiAction("", "png|PNG|jpg|JPG|jpeg|mp4|m4v|mov|MOV|webm")), Hand: func(m *ice.Message, arg ...string) {
			_wiki_list(m.Options(nfs.DIR_ROOT, _feel_path(m, "")), kit.Slice(arg, 0, 1)...)
			ctx.DisplayLocal(m, "")
		}},
	})
}
