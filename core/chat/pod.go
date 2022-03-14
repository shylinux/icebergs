package chat

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
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
			if kit.Select("", arg, 0) == "" {
				m.RenderCmd(web.ROUTE)
				return // 节点列表
			}
			if len(arg) == 1 {
				if s := m.Cmdx(web.SPACE, arg[0], "web.chat.website", "show", "index.iml", "Header", "", "River", "", "Action", "", "Footer", ""); s != "" {
					m.RenderResult(s)
				} else if s := m.Cmdx(web.SPACE, arg[0], nfs.CAT, "/page/index.html"); s != "" {
					m.RenderResult(s)
				} else {
					m.RenderIndex(web.SERVE, ice.VOLCANOS)
				}
				return // 节点首页
			}
			if arg[1] == WEBSITE {
				m.Cmdy(web.SPACE, arg[0], WEBSITE, ctx.ACTION, "show", path.Join(arg[2:]...))
				m.RenderResult()
				return
			}
			// 节点命令
			m.Cmdy("/cmd/", path.Join(arg[2:]...))
		}},
	}})
}
