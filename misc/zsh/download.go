package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{},
		Commands: map[string]*ice.Command{
			"/download": {Name: "/download", Help: "下载", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 || arg[0] == "" {
					// 文件列表
					m.Cmdy(web.SPACE, m.Option("you"), web.STORY).Table()
					return
				}

				// 查找文件
				if m.Cmdy(web.STORY, "index", arg[0]).Append("text") == "" && m.Option("you") != "" {
					// 上发文件
					m.Cmd(web.SPACE, m.Option("you"), web.STORY, "push", arg[0], "dev", arg[0])
					m.Cmdy(web.STORY, "index", arg[0])
				}

				// 下载文件
				m.Render(kit.Select(ice.RENDER_DOWNLOAD, ice.RENDER_RESULT, m.Append("file") == ""), m.Append("text"))
			}},
			"/upload": {Name: "/upload", Help: "上传", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				// 缓存文件
				msg := m.Cmd(web.STORY, "upload")
				m.Echo("data: %s\n", msg.Append("data"))
				m.Echo("time: %s\n", msg.Append("time"))
				m.Echo("type: %s\n", msg.Append("type"))
				m.Echo("name: %s\n", msg.Append("name"))
				m.Echo("size: %s\n", msg.Append("size"))
			}},
		},
	})
}
