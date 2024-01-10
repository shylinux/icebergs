package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const PORTAL = "portal"

func init() {
	Index.MergeCommands(ice.Commands{
		PORTAL: {Help: "门户", Actions: ice.MergeActions(ice.Actions{
			web.DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) { m.PushButton(kit.Dict(web.ADMIN, "后台")) }},
			web.DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				if kit.HasPrefixList(arg, ctx.ACTION, web.ADMIN) && len(arg) == 2 {
					ctx.ProcessField(m, web.CHAT_IFRAME, m.MergePodCmd(m.Option(mdb.NAME), m.PrefixKey()), arg...)
					m.ProcessField(ctx.ACTION, ctx.RUN, web.CHAT_IFRAME)
				}
			}},
		}, web.ApiAction()), Hand: func(m *ice.Message, arg ...string) { web.RenderMain(m) }},
	})
}
