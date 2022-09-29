package chat

import (
	"net/http"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const MEDIA = "media"

func init() {
	Index.MergeCommands(ice.Commands{
		MEDIA: {Name: "media path auto", Help: "媒体", Actions: ice.MergeActions(ice.Actions{
			web.UPLOAD: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(web.CACHE, web.UPLOAD_WATCH, ice.USR_LOCAL_MEDIA)
			}},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.TRASH, path.Join(ice.USR_LOCAL_MEDIA, m.Option(nfs.PATH)))
			}},
		}, web.ApiAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Option(nfs.DIR_ROOT, ice.USR_LOCAL_MEDIA)
				ctx.DisplayStory(m.Cmdy(nfs.DIR, nfs.PWD), "media.js")
				return
			}
			if m.R.Method == http.MethodGet {
				m.RenderDownload(kit.Path(ice.USR_LOCAL_MEDIA, path.Join(arg...)))
			} else {
				m.EchoVideos("/chat/media/"+path.Join(arg...), m.Option("height"))
			}
		}},
	})
}
