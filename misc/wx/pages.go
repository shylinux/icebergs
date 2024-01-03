package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

func init() {
	web.Index.MergeCommands(ice.Commands{
		web.PP(PAGES): {Actions: aaa.WhiteAction("", ctx.ACTION), Hand: func(m *ice.Message, arg ...string) {
			if len(arg[0]) == 0 || arg[0] == "" || arg[0] == chat.RIVER {
				web.RenderMain(m)
			} else {
				if m.IsWeixinUA() {
				}
				web.RenderCmd(m, kit.Select(m.Option(ctx.INDEX), m.Option(ice.CMD)))
			}
		}},
	})
}
