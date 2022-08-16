package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const TEMPLATE = "template"

func init() {
	Index.MergeCommands(ice.Commands{
		TEMPLATE: {Name: "template river storm index auto 删除配置 查看配置", Help: "模板", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				if gdb.Watch(m, RIVER_CREATE); m.Cmd("").Length() == 0 {
					kit.Fetch(_river_template, func(river string, value ice.Any) {
						m.Cmd("", mdb.CREATE, RIVER, river)
						kit.Fetch(value, func(storm string, value ice.Any) {
							m.Cmd("", mdb.INSERT, RIVER, river, mdb.TYPE, "", STORM, storm, mdb.TEXT, "")
							kit.Fetch(value, func(index int, value ice.Any) {
								m.Cmd("", "add", RIVER, river, STORM, storm, ctx.INDEX, value)
							})
						})
					})
				}
			}},
			RIVER_CREATE: {Name: "river.create river template", Help: "建群", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", m.Option(TEMPLATE), func(value ice.Maps) {
					h := m.Cmdx(STORM, mdb.CREATE, mdb.TYPE, "", mdb.NAME, value[STORM], mdb.TEXT, "")
					m.Cmd("", m.Option(TEMPLATE), value[STORM], func(value ice.Maps) {
						m.Cmd(STORM, mdb.INSERT, mdb.HASH, h, kit.SimpleKV("space,index", value))
					})
				})
			}},
			mdb.CREATE: {Name: "create river", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(RIVER), kit.Dict(mdb.SHORT, RIVER))
			}},
			mdb.INSERT: {Name: "insert river type storm text", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), kit.KeyHash(m.Option(RIVER)), mdb.HASH, arg[2:], kit.Dict(mdb.SHORT, STORM))
			}},
			"add": {Name: "add river storm index", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), kit.KeyHash(m.Option(RIVER), kit.KeyHash(m.Option(STORM))), mdb.LIST, arg[4:])
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(STORM) == "" {
					m.Cmd(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(RIVER))
				} else {
					m.Cmd(mdb.DELETE, m.PrefixKey(), kit.KeyHash(m.Option(RIVER)), mdb.HASH, m.OptionSimple(STORM))
				}
			}},
		}), Hand: func(m *ice.Message, arg ...string) {
			switch len(arg) {
			case 0:
				m.OptionFields("time,river")
				m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.HASH)
				m.PushAction(mdb.REMOVE)
				m.Action(mdb.CREATE)
			case 1:
				m.OptionFields("time,type,storm,text")
				m.Cmdy(mdb.SELECT, m.PrefixKey(), kit.KeyHash(arg[0]), mdb.HASH)
				m.PushAction(mdb.REMOVE)
				m.Action(mdb.INSERT)
			case 2:
				m.OptionFields("time,index")
				m.Cmdy(mdb.SELECT, m.PrefixKey(), kit.KeyHash(arg[0], kit.KeyHash(arg[1])), mdb.LIST)
				m.Action("add")
			}
		}},
	})
}

var _river_template = kit.Dict(
	"base", kit.Dict(
		"info", kit.List(
			"web.chat.storm",
			"web.chat.ocean",
			"web.chat.nodes",
		),
		"scan", kit.List(
			"web.chat.scan",
			"web.chat.paste",
			"web.chat.files",
			"web.chat.location",
			"web.chat.meet.miss",
			"web.wiki.feel",
		),
		"task", kit.List(
			"web.team.task",
			"web.team.plan",
			"web.mall.asset",
			"web.mall.salary",
			"web.wiki.word",
		),
		"draw", kit.List(
			"web.wiki.draw",
			"web.wiki.data",
			"web.wiki.word",
		),
		"term", kit.List(
			"web.code.xterm",
			"web.code.vimer",
		),
	),
)
