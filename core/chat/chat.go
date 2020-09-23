package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const CHAT = "chat"

var Index = &ice.Context{Name: CHAT, Help: "聊天中心",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Conf(RIVER, "meta.template", kit.Dict(
				"base", kit.Dict(
					"info", []interface{}{
						"web.chat.info",
						"web.chat.auth",
						"web.chat.user",
						"web.chat.tool",
						"web.chat.node",
					},
					"miss", []interface{}{
						"web.team.task",
						"web.team.plan",
						"web.wiki.word",
					},
				),
			))
			m.Watch(web.SPACE_START, m.Prefix(NODE))
			m.Watch(web.SPACE_CLOSE, m.Prefix(NODE))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save(RIVER)
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.Frame{},
		HEADER, RIVER, STORM, ACTION, FOOTER,
	)
}
