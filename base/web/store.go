package web

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const STORE = "store"

func init() {
	Index.MergeCommands(ice.Commands{
		STORE: {Help: "系统商店", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmds(SPIDE).Table(func(value ice.Maps) {
					if value[CLIENT_TYPE] == nfs.REPOS {
						m.Push("", value, arg[0])
					}
				})
			}},
			"install": {Help: "购买", Hand: func(m *ice.Message, arg ...string) {
				if strings.HasPrefix(m.Option(mdb.ICON), "/require/") {
					m.Option(mdb.ICON, strings.TrimSuffix(strings.TrimPrefix(m.Option(mdb.ICON), "/require/"), "?pod="+m.Option(mdb.NAME)))
				}
				m.OptionDefault(nfs.BINARY, m.Option(ORIGIN)+S(m.Option(mdb.NAME)))
				m.Cmdy(DREAM, mdb.CREATE, m.OptionSimple(mdb.NAME, mdb.ICON, nfs.REPOS, nfs.BINARY))
				m.Cmdy(DREAM, cli.START, m.OptionSimple(mdb.NAME))
			}},
			PORTAL: {Help: "详情", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, CHAT_IFRAME, m.Option(ORIGIN)+S(m.Option(mdb.NAME))+C(PORTAL), arg...)
			}},
			mdb.CREATE: {Name: "create name* origin*", Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.TYPE, nfs.REPOS)
				m.Cmd(SPIDE, mdb.CREATE, m.OptionSimple("name,origin,type"))
			}},
		}, ctx.ConfAction(ctx.TOOLS, "dream")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(SPIDE, arg, kit.Dict(ice.MSG_FIELDS, "time,client.type,client.name,client.origin")).Action(mdb.CREATE).Display("")
				ctx.Toolkit(m)
			} else {
				origin := SpideOrigin(m, arg[0])
				m.SetAppend().SplitIndex(m.Cmdx(SPIDE, arg[0], C(DREAM))).Table(func(value ice.Maps) { m.Push(ORIGIN, origin) }).PushAction("install", PORTAL)
			}
		}},
	})
}
