package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/download": {Name: "/download", Help: "下载", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 || arg[0] == "" {
				m.Cmdy(chat.FILES).Table()
				return // 文件列表
			}

			// 下载文件
			m.Cmdy(web.CACHE, m.Cmd(chat.FILES, arg[0]).Append(kit.MDB_DATA))
			m.Render(kit.Select(ice.RENDER_DOWNLOAD, ice.RENDER_RESULT, m.Append(kit.MDB_FILE) == ""), m.Append(kit.MDB_TEXT))
		}},
		"/upload": {Name: "/upload", Help: "上传", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 上传文件
			msg := m.Cmd(web.CACHE, web.UPLOAD)
			m.Option(ice.MSG_UPLOAD, msg.Append(kit.MDB_HASH), msg.Append(kit.MDB_NAME), msg.Append(kit.MDB_SIZE))
			m.Cmd(chat.FILES, web.UPLOAD)

			for _, k := range []string{kit.MDB_DATA, kit.MDB_TIME, kit.MDB_TYPE, kit.MDB_NAME, kit.MDB_SIZE} {
				m.Echo("%s: %s\n", k, msg.Append(k))
			}
		}},
	}})
}
