package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/web"
)

const CMD = "cmd"

func init() {
	Index.MergeCommands(ice.Commands{
		CMD: {Actions: ice.MergeActions(web.ApiAction(), aaa.WhiteAction(ctx.RUN), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg[0]) == 0 || arg[0] == "" {
				web.RenderMain(m)
			} else if m.IsCliUA() {
				if aaa.Right(m, arg) {
					m.Cmdy(arg, m.Optionv(ice.ARG)).RenderResult()
				}
			} else if arg[0] == web.CHAT_PORTAL {
				web.RenderMain(m)
			} else if m.Cmdy(ctx.COMMAND, arg[0]); m.Length() > 0 {
				web.RenderCmd(m, m.Append(ctx.INDEX), arg[1:])
			}
		}},
	})
}
