package wiki

import (
	"path"
	"strings"

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
	return p
	if nfs.Exists(m, ice.USR_LOCAL_IMAGE) {
		return path.Join(ice.USR_LOCAL_IMAGE, p)
	}
	return p
}

const FEEL = "feel"

func init() {
	Index.MergeCommands(ice.Commands{
		FEEL: {Name: "feel path=usr/icons/@key file=background.jpg auto upload record1 record2 actions", Icon: "Photos.png", Help: "影音媒体", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				m.Push(arg[0], "usr/icons/")
				m.Push(arg[0], "usr/local/image/")
			}},
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
				m.Cmdy(web.CACHE, web.WATCH, m.Option(mdb.HASH), path.Join(m.Option(nfs.PATH, _feel_path(m, m.Option(nfs.PATH))), up[1]))
			}},
			"moveto": {Hand: func(m *ice.Message, arg ...string) {
				kit.For(arg[1:], func(from string) { m.Cmd(nfs.MOVE, path.Join(arg[0], path.Base(from)), from) })
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				p := kit.Select(_feel_path(m, m.Option(nfs.PATH)), arg, 0)
				kit.If(strings.HasSuffix(p, nfs.PS), func() { mdb.HashRemove(m, nfs.PATH, p) })
				nfs.Trash(m, p)
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
		}, mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,name,path,cover"), chat.FavorAction(), WikiAction("", "ico|png|PNG|jpg|JPG|jpeg|mp4|m4v|mov|MOV|webm")), Hand: func(m *ice.Message, arg ...string) {
			if len(kit.Slice(arg, 0, 2)) == 0 {
				mdb.HashSelect(m)
				m.Push(nfs.PATH, "usr/image/").Push(mdb.NAME, "照片库").Push("cover", "usr/icons/background.jpg")
				m.Push(nfs.PATH, "usr/avatar/").Push(mdb.NAME, "头像库").Push("cover", "usr/icons/avatar.jpg")
				m.Push(nfs.PATH, "usr/icons/").Push(mdb.NAME, "图标库").Push("cover", "src/main.ico")
				return
			}
			_wiki_list(m.Options(nfs.DIR_ROOT, _feel_path(m, "")), kit.Slice(arg, 0, 1)...)
			m.SortStrR(mdb.TIME)
			ctx.DisplayLocal(m, "")
		}},
	})
}
