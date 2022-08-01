package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.MergeCommands(ice.Commands{
		"/download": {Name: "/download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				m.Cmdy(chat.FILES).Table()
				return // 文件列表
			}

			// 下载文件
			m.Cmdy(web.CACHE, m.Cmd(chat.FILES, arg[0]).Append(mdb.DATA))
			m.Render(kit.Select(ice.RENDER_DOWNLOAD, ice.RENDER_RESULT, m.Append(nfs.FILE) == ""), m.Append(mdb.TEXT))
		}},
		"/upload": {Name: "/upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
			msg := m.Cmd(chat.FILES, web.UPLOAD) // 上传文件
			for _, k := range []string{mdb.DATA, mdb.TIME, mdb.TYPE, mdb.NAME, nfs.SIZE} {
				m.Echo("%s: %s\n", k, msg.Append(k))
			}
		}},
	})
}
