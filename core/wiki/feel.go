package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

const (
	USR_LOCAL_IMAGE      = "usr/local/image/"
	USR_IMAGE            = "usr/image/"
	USR_COVER            = "usr/cover/"
	USR_AVATAR           = "usr/avatar/"
	USR_ICONS            = "usr/icons/"
	USR_ICONS_AVATAR     = "usr/icons/avatar.jpg"
	USR_ICONS_BACKGROUND = "usr/icons/background.jpg"
	SRC_MAIN             = "src/main.ico"
)
const (
	COVER = "cover"
)
const FEEL = "feel"

func init() {
	Index.MergeCommands(ice.Commands{
		FEEL: {Name: "feel path=usr/icons/@key file=background.jpg auto", Help: "影音媒体", Icon: "Photos.png", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy("").Cut(nfs.PATH)
			}},
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				up := kit.Simple(m.Optionv(ice.MSG_UPLOAD))
				m.Cmdy(web.CACHE, web.WATCH, m.Option(mdb.HASH), path.Join(m.Option(nfs.PATH), up[1]))
			}},
			nfs.MOVETO: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.MOVETO, arg)
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				p := kit.Select(m.Option(nfs.PATH), arg, 0)
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
		}, chat.FavorAction(), WikiAction("", "ico|png|PNG|jpg|JPG|jpeg|mp4|m4v|mov|MOV|webm|mp3"), mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path,name,cover")), Hand: func(m *ice.Message, arg ...string) {
			if len(kit.Slice(arg, 0, 1)) == 0 {
				if mdb.HashSelect(m); aaa.IsTechOrRoot(m) {
					m.Push(nfs.PATH, USR_AVATAR).Push(mdb.NAME, "头像库").Push(COVER, USR_ICONS_AVATAR)
					m.Push(nfs.PATH, USR_LOCAL_IMAGE).Push(mdb.NAME, "私有库").Push(COVER, USR_ICONS_BACKGROUND)
				}
				m.Push(nfs.PATH, USR_IMAGE).Push(mdb.NAME, "照片库").Push(COVER, USR_ICONS_BACKGROUND)
				m.Push(nfs.PATH, USR_COVER).Push(mdb.NAME, "封面库").Push(COVER, USR_ICONS_BACKGROUND)
				m.Push(nfs.PATH, USR_ICONS).Push(mdb.NAME, "图标库").Push(COVER, SRC_MAIN)
			} else {
				if _wiki_list(m, kit.Slice(arg, 0, 1)...); arg[0] == USR_ICONS {
					m.Sort(mdb.NAME)
				} else {
					switch kit.Select(mdb.TIME, arg, 2) {
					case mdb.TIME:
						m.SortStrR(mdb.TIME)
					case nfs.PATH:
						m.Sort(nfs.PATH)
					case nfs.SIZE:
						m.SortIntR(nfs.SIZE)
					}
				}
				list := m.Spawn().Options(nfs.DIR_DEEP, ice.TRUE).CmdMap(nfs.DIR, USR_COVER+arg[0], nfs.PATH)
				m.Table(func(value ice.Maps) {
					p := USR_COVER + kit.TrimSuffix(value[nfs.PATH], ".mp3", ".mp4") + ".jpg"
					if _, ok := list[p]; ok {
						m.Push(COVER, p)
					} else {
						m.Push(COVER, "")
					}
				})
			}
			ctx.DisplayLocal(m, "")
		}},
	})
}
