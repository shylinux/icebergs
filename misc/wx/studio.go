package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const STUDIO = "studio"

func init() {
	Index.MergeCommands(ice.Commands{
		STUDIO: {Icon: "wxdev.png", Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(ctx.COMMAND, kit.Split(kit.Select("web.chat.wx.access,web.chat.wx.ide,web.chat.wx.scan", mdb.Config(m, ctx.CMDS))))
			ctx.DisplayStory(m, "")
		}},
	})
}
