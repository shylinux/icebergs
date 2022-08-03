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
	Index.MergeCommands(ice.Commands{
		"/pod/": {Name: "/pod/", Help: "节点", Actions: ctx.CmdAction(), Hand: func(m *ice.Message, arg ...string) {
			if strings.HasPrefix(m.R.Header.Get("User-Agent"), "curl") || strings.HasPrefix(m.R.Header.Get("User-Agent"), "Wget") {
				m.Option(ice.MSG_USERNAME, "root")
				m.Option(ice.MSG_USERROLE, "root")
				m.Option(ice.POD, kit.Select("", arg, 0))
				m.Cmdy(web.SHARE_LOCAL, "bin/ice.bin")
				return // 下载文件
			}

			if len(arg) == 0 || kit.Select("", arg, 0) == "" { // 节点列表
				web.RenderCmd(m, web.ROUTE)

			} else if len(arg) == 1 { // 节点首页
				if m.Cmd(web.SPACE, arg[0]).Length() == 0 && !strings.Contains(arg[0], ice.PT) {
					m.Cmd(web.DREAM, cli.START, mdb.NAME, arg[0])
				}
				aaa.UserRoot(m)
				if web.RenderWebsite(m, arg[0], "index.iml", "Header", "", "River", "", "Footer", ""); m.Result() == "" {
					web.RenderIndex(m, web.SERVE, ice.VOLCANOS)
				}

			} else if arg[1] == WEBSITE { // 节点网页
				web.RenderWebsite(m, arg[0], path.Join(arg[2:]...))

			} else if arg[1] == "cmd" { // 节点命令
				m.Cmdy(web.SPACE, arg[0], m.Prefix(CMD), path.Join(arg[2:]...))
			} else {
				m.Cmdy(web.SPACE, m.Option(ice.MSG_USERPOD), "web.chat."+strings.TrimPrefix(path.Join(arg[1:]...), "chat/"))
				// m.Cmdy(web.SPACE, m.Option(ice.MSG_USERPOD), "web.chat."+ice.PS+path.Join(arg[1:]...))
			}
		}},
	})
}
