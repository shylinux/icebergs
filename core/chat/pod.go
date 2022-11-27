package chat

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const POD = "pod"

func init() {
	Index.MergeCommands(ice.Commands{
		POD: {Name: "pod", Help: "节点", Actions: ice.MergeActions(ice.Actions{
			web.SERVE_PARSE: {Hand: func(m *ice.Message, arg ...string) {
				switch kit.Select("", arg, 0) {
				case CHAT:
					for i := 1; i < len(arg)-1; i++ {
						m.Logs("refer", arg[i], arg[i+1])
						m.Option(arg[i], arg[i+1])
					}
				}
			}},
		}, ctx.CmdAction(), web.ServeAction(), web.ApiAction(), aaa.WhiteAction()), Hand: func(m *ice.Message, arg ...string) {
			if web.OptionAgentIs(m, "curl", "wget") {
				aaa.UserRoot(m).Cmdy(web.SHARE_LOCAL, ice.BIN_ICE_BIN, kit.Dict(ice.POD, kit.Select("", arg, 0)))
				return
			}
			if len(arg) == 0 || kit.Select("", arg, 0) == "" {
				web.RenderCmd(m, web.SPACE)
			} else if len(arg) == 1 {
				if m.Cmd(web.SPACE, arg[0]).Length() == 0 && !strings.Contains(arg[0], ice.PT) {
					m.Cmd(web.DREAM, cli.START, mdb.NAME, arg[0])
				}
				web.RenderMain(aaa.UserRoot(m), arg[0], "")
			} else if arg[1] == CMD {
				web.RenderPodCmd(m, arg[0], arg[2], arg[3:])
			} else if arg[1] == WEBSITE {
				RenderWebsite(m, arg[0], path.Join(arg[2:]...))
			}
		}},
	})
}

func RenderWebsite(m *ice.Message, pod string, dir string, arg ...string) *ice.Message {
	return m.Echo(m.Cmdx(web.Space(m, pod), "web.chat.website", lex.PARSE, dir, arg)).RenderResult()
}
