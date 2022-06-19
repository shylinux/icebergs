package chat

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
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
			if strings.HasPrefix(m.R.Header.Get("User-Agent"), "curl") || strings.HasPrefix(m.R.Header.Get("User-Agent"), "Wget") {
				m.Option(ice.MSG_USERNAME, "root")
				m.Option(ice.MSG_USERROLE, "root")
				m.Option(ice.POD, kit.Select("", arg, 0))
				m.Cmdy(web.SHARE_LOCAL, "bin/ice.bin")
				return // 下载文件
			}

			if len(arg) == 0 || kit.Select("", arg, 0) == "" { // 节点列表
				m.RenderCmd(web.ROUTE)

			} else if len(arg) == 1 { // 节点首页
				if m.Cmd(web.SPACE, arg[0]).Length() == 0 {
					m.Cmd(web.DREAM, cli.START, mdb.NAME, arg[0])
				}
				aaa.UserRoot(m)
				if m.RenderWebsite(arg[0], "index.iml", "Header", "", "River", "", "Action", "", "Footer", ""); m.Result() == "" {
					m.RenderIndex(web.SERVE, ice.VOLCANOS)
				}

			} else if arg[1] == WEBSITE { // 节点网页
				m.RenderWebsite(arg[0], path.Join(arg[2:]...))

			} else if arg[1] == "cmd" { // 节点命令
				m.Cmdy(web.SPACE, arg[0], m.Prefix(CMD), path.Join(arg[2:]...))
			} else {
				m.Debug("what %v", path.Join(arg[1:]...))
				m.Cmdy(web.SPACE, m.Option(ice.MSG_USERPOD), "web.chat."+strings.TrimPrefix(path.Join(arg[1:]...), "chat/"))
				// m.Cmdy(web.SPACE, m.Option(ice.MSG_USERPOD), "web.chat."+ice.PS+path.Join(arg[1:]...))
			}
		}},
	}})
}
