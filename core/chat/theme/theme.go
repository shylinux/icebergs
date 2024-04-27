package theme

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

func init() {
	const THEME = "theme"
	chat.Index.MergeCommands(ice.Commands{
		THEME: {Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy("web.chat.color").Cut(mdb.NAME, mdb.TEXT, mdb.HELP) }},
			mdb.CREATE: {Name: "create name* plugin-bg-color@color output-bg-color@color hover-bg-color@color"},
			mdb.SHOW: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessFloat(m, "web.chat.iframe", []string{"/?theme=" + m.Option(mdb.NAME)}, arg...)
			}},
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) == 0 {
					m.Cmdy("web.chat.theme")
				} else {
					m.Cmdy("web.chat.theme", kit.TrimExt(arg[0], nfs.CSS)).RenderResult()
					web.RenderType(m.W, arg[0], "")
				}
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, kit.Fields("time,name",
			"plugin-bg-color", "output-bg-color", "hover-bg-color",
			"plugin-fg-color", "output-fg-color", "hover-fg-color",
			"shadow-color", "border-color", "notice-color", "danger-color",
		))), Hand: func(m *ice.Message, arg ...string) {
			if mdb.HashSelect(m, arg...).PushAction(mdb.SHOW, mdb.REMOVE).Display(""); len(arg) > 0 {
				defer m.Echo("body.%s {"+lex.NL, kit.TrimExt(arg[0], nfs.CSS)).Echo("}" + lex.NL)
				m.Table(func(value ice.Maps) {
					kit.For(value, func(k, v string) {
						if v == "" || kit.IsIn(k, mdb.TIME, mdb.NAME, ctx.ACTION) {
							return
						}
						m.Echo(kit.Format("\t--%s: %s;\n", k, v))
					})
				})
			}
		}},
	})
}
