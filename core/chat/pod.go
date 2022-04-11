package chat

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const POD = "pod"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		POD: {Name: POD, Help: "节点", Value: kit.Data()},
	}, Commands: map[string]*ice.Command{
		"/pod/": {Name: "/pod/", Help: "节点", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Name: "_init", Help: "初始化", Hand: func(m *ice.Message, arg ...string) {
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.IsCliUA() {
				m.Option(ice.MSG_USERNAME, "root")
				m.Option(ice.MSG_USERROLE, "root")
				m.Option(ice.POD, kit.Select("", arg, 0))
				m.Cmdy(web.SHARE_LOCAL, "bin/ice.bin")
				return // 下载文件
			}

			if len(arg) == 0 || kit.Select("", arg, 0) == "" { // 节点列表
				m.RenderCmd(web.ROUTE)

			} else if len(arg) == 1 { // 节点首页
				aaa.UserRoot(m)
				if m.RenderWebsite(arg[0], "index.iml", "Header", "", "River", "", "Action", "", "Footer", ""); m.Result() == "" {
					m.RenderIndex(web.SERVE, ice.VOLCANOS)
				}

			} else if arg[1] == WEBSITE { // 节点网页
				m.RenderWebsite(arg[0], path.Join(arg[2:]...))

			} else { // 节点命令
				m.Cmdy("/cmd/", path.Join(arg[2:]...))
			}
		}},
	}})
}
