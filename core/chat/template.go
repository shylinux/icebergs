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
							m.Cmd("", mdb.INSERT, RIVER, river, STORM, storm)
							kit.Fetch(value, func(index int, value ice.Any) {
								m.Cmd("", "add", RIVER, river, STORM, storm, ctx.INDEX, value)
							})
						})
					})
				}
			}},
			RIVER_CREATE: {Name: "river.create river template", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", m.Option(TEMPLATE), ice.OptionFields(STORM), func(value ice.Maps) {
					m.Option(ice.MSG_STORM, m.Cmdx(STORM, mdb.CREATE, mdb.NAME, value[STORM]))
					m.Cmd("", m.Option(TEMPLATE), value[STORM], ice.OptionFields(ctx.INDEX), func(value ice.Maps) {
						m.Cmd(STORM, mdb.INSERT, ctx.INDEX, value[ctx.INDEX])
					})
				})
			}},
			mdb.CREATE: {Name: "create river", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(RIVER), kit.Dict(mdb.SHORT, RIVER))
			}},
			mdb.INSERT: {Name: "insert river storm", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), kit.KeyHash(m.Option(RIVER)), mdb.HASH, arg[2:], kit.Dict(mdb.SHORT, STORM))
			}},
			"add": {Name: "add river storm index", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), kit.KeyHash(m.Option(RIVER), kit.KeyHash(m.Option(STORM))), mdb.LIST, arg[4:])
			}},
			mdb.REMOVE: {Name: "remove", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(STORM) == "" {
					m.Cmd(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(RIVER))
				} else {
					m.Cmd(mdb.DELETE, m.PrefixKey(), kit.KeyHash(m.Option(RIVER)), mdb.HASH, m.OptionSimple(STORM))
				}
			}},
		}, mdb.ClearHashOnExitAction()), Hand: func(m *ice.Message, arg ...string) {
			switch len(arg) {
			case 0:
				m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, ice.OptionFields("time,river"))
				m.PushAction(mdb.REMOVE).Action(mdb.CREATE)
			case 1:
				m.Cmdy(mdb.SELECT, m.PrefixKey(), kit.KeyHash(arg[0]), mdb.HASH, ice.OptionFields("time,storm"))
				m.PushAction(mdb.REMOVE).Action(mdb.INSERT)
			case 2:
				m.Cmdy(mdb.SELECT, m.PrefixKey(), kit.KeyHash(arg[0], kit.KeyHash(arg[1])), mdb.LIST, ice.OptionFields("time,index"))
				m.Action("add")
			}
		}},
	})
}

var _river_template = kit.Dict(
	"base", kit.Dict(
		"draw", kit.List(
			"web.wiki.draw",
			"web.wiki.data",
			"web.wiki.word",
		),
		"term", kit.List(
			"web.code.xterm",
			"web.code.vimer",
			"web.chat.iframe",
		),
		"task", kit.List(
			"web.team.task",
			"web.team.plan",
			"web.mall.asset",
			"web.mall.salary",
			"web.wiki.word",
		),
		"scan", kit.List(
			"web.chat.scan",
			"web.chat.paste",
			"web.chat.files",
			"web.chat.location",
			"web.chat.meet.miss",
			"web.wiki.feel",
		),
	),
)
