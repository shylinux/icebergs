package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const CHAT = "chat"

var Index = &ice.Context{Name: CHAT, Help: "聊天中心",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(web.SERVE, aaa.WHITE, HEADER, RIVER, ACTION, FOOTER)
			m.Cmd(mdb.SEARCH, mdb.CREATE, P_SEARCH, m.Prefix(P_SEARCH))
			m.Cmd(mdb.SEARCH, mdb.CREATE, EMAIL, m.Prefix(EMAIL))
			m.Watch(web.SPACE_START, m.Prefix(NODE))
			m.Watch(web.SPACE_STOP, m.Prefix(NODE))
			m.Load()

			for _, cmd := range []string{
				"web.chat.meet.miss",
				"web.chat.meet.mate",
				"web.chat.location",
				"web.chat.paste",
				"web.chat.scan",
				"web.wiki.feel",
				"web.wiki.draw",
				"web.wiki.data",
				"web.wiki.word",
				"web.team.task",
				"web.team.plan",
				"web.mall.asset",
				"web.mall.salary",
			} {
				m.Conf(ACTION, kit.Keym("domain", cmd), "true")
			}

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
						"web.chat.meet.miss",
						"web.wiki.feel",
					},
					"task", []interface{}{
						"web.team.task",
						"web.team.plan",
						"web.mall.asset",
						"web.mall.salary",
						"web.wiki.word",
					},
					"draw", []interface{}{
						"web.wiki.draw",
						"web.wiki.data",
						"web.wiki.word",
					},
				),
			))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},
	},
}

func init() {
	web.Index.Register(Index, &web.Frame{},
		HEADER, RIVER, STORM, ACTION, FOOTER,
		SCAN, PASTE, FILES, LOCATION,
	)
}
