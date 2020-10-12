package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const CHAT = "chat"

var Index = &ice.Context{Name: CHAT, Help: "聊天中心",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(web.SERVE, aaa.WHITE, "header", "river", "action", "footer")

			m.Conf(ACTION, "meta.domain.web.chat.location", "true")
			m.Conf(ACTION, "meta.domain.web.chat.paste", "true")
			m.Conf(ACTION, "meta.domain.web.chat.scan", "true")
			m.Conf(ACTION, "meta.domain.web.wiki.feel", "true")

			m.Conf(RIVER, "meta.template", kit.Dict(
				"base", kit.Dict(
					"info", []interface{}{
						"web.chat.info",
						"web.chat.auth",
						"web.chat.user",
						"web.chat.tool",
						"web.chat.node",
					},
					"scan", []interface{}{
						"web.chat.scan",
						"web.chat.paste",
						"web.chat.location",
					},
					"miss", []interface{}{
						"web.team.task",
						"web.team.plan",
						"web.wiki.draw",
						"web.wiki.data",
						"web.wiki.word",
					},
					"meet", []interface{}{
						"web.wiki.feel",
						"web.chat.meet.miss",
						"web.wiki.word",
					},
				),
			))
			m.Watch(web.SPACE_START, m.Prefix(NODE))
			m.Watch(web.SPACE_STOP, m.Prefix(NODE))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},
	},
}

func init() {
	web.Index.Register(Index, &web.Frame{},
		HEADER, RIVER, STORM, ACTION, FOOTER,
	)
}
