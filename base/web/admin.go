package web

import (
	"net/http"
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const ADMIN = "admin"

func init() {
	Index.MergeCommands(ice.Commands{
		ADMIN: {Name: "admin index list", Help: "管理", Actions: ice.Actions{
			DREAM_TABLES: {Hand: func(m *ice.Message, arg ...string) { m.PushButton(kit.Dict(ADMIN, "后台")) }},
			DREAM_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				if kit.HasPrefixList(arg, ctx.ACTION, ADMIN) && len(arg) == 2 {
					if m.Option(mdb.TYPE) == MASTER {
						// m.ProcessOpen(SpideOrigin(m, m.Option(mdb.NAME)) + C(m.PrefixKey()))
						ctx.ProcessField(m, CHAT_IFRAME, SpideOrigin(m, m.Option(mdb.NAME))+C(m.PrefixKey()), arg...)
						m.ProcessField(ctx.ACTION, ctx.RUN, CHAT_IFRAME)
					} else {
						ctx.ProcessField(m, CHAT_IFRAME, m.MergePodCmd(m.Option(mdb.NAME), ""), arg...)
						m.ProcessField(ctx.ACTION, ctx.RUN, CHAT_IFRAME)
					}
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(ice.MSG_SOURCE) != "" {
				RenderMain(m)
			} else {
				kit.If(len(arg) == 0, func() { arg = append(arg, SPACE, DOMAIN) })
				m.Cmdy(SPIDE, ice.OPS, SPIDE_RAW, http.MethodPost, C(path.Join(arg...)), "pwd", kit.Path(""))
			}
		}},
	})
}
func AdminCmd(m *ice.Message, cmd string, arg ...string) string {
	if ice.Info.NodeType == WORKER {
		return m.Cmdx(SPIDE, ice.OPS, SPIDE_RAW, http.MethodGet, path.Join(C(cmd), path.Join(arg...)))
	} else {
		return m.Cmdx(cmd, arg)
	}
}
