package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			"/download": {Name: "/download", Help: "下载", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 || arg[0] == "" {
					m.Cmdy("web.chat.files").Table()
					return
				}

				// 下载文件
				m.Cmdy(web.CACHE, m.Cmd("web.chat.files", arg[0]).Append("data"))
				m.Render(kit.Select(ice.RENDER_DOWNLOAD, ice.RENDER_RESULT, m.Append("file") == ""), m.Append("text"))
			}},
			"/upload": {Name: "/upload", Help: "上传", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				// 缓存文件
				msg := m.Cmd(web.CACHE, web.UPLOAD)
				m.Option(ice.MSG_UPLOAD, msg.Append(kit.MDB_HASH), msg.Append(kit.MDB_NAME))
				m.Cmd("web.chat.files", "upload")

				m.Echo("data: %s\n", msg.Append("data"))
				m.Echo("time: %s\n", msg.Append("time"))
				m.Echo("type: %s\n", msg.Append("type"))
				m.Echo("name: %s\n", msg.Append("name"))
				m.Echo("size: %s\n", msg.Append("size"))
			}},
		},
	})
}
