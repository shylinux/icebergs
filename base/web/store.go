package web

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const STORE = "store"

func init() {
	Index.MergeCommands(ice.Commands{
		STORE: {Name: "store list", Help: "系统商店", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmds(SPIDE).Table(func(value ice.Maps) { kit.If(value[CLIENT_TYPE] == nfs.REPOS, func() { m.Push("", value, arg[0]) }) })
			}},
			mdb.CREATE: {Name: "create name* origin*", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPIDE, mdb.CREATE, m.OptionSimple("name,origin"), mdb.TYPE, nfs.REPOS)
			}},
			INSTALL: {Hand: func(m *ice.Message, arg ...string) {
				if !kit.HasPrefixList(arg, ctx.RUN) {
					if !nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME))) {
						if strings.HasPrefix(m.Option(mdb.ICON), nfs.REQUIRE) {
							m.Option(mdb.ICON, strings.TrimSuffix(strings.TrimPrefix(m.Option(mdb.ICON), nfs.REQUIRE), "?pod="+m.Option(mdb.NAME)))
						}
						m.OptionDefault(nfs.BINARY, m.Option(ORIGIN)+S(m.Option(mdb.NAME)))
						m.Cmdy(DREAM, mdb.CREATE, m.OptionSimple(mdb.NAME, mdb.ICON, nfs.REPOS, nfs.BINARY))
						m.Cmdy(DREAM, cli.START, m.OptionSimple(mdb.NAME))
					}
					defer m.Push("title", m.Option(mdb.NAME))
				}
				ctx.ProcessField(m, CHAT_IFRAME, S(m.Option(mdb.NAME)), arg...)
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, CHAT_IFRAME, S(m.Option(mdb.NAME)), arg...)
				kit.If(!kit.HasPrefixList(arg, ctx.RUN), func() { m.Push("title", m.Option(mdb.NAME)) })
			}},
			PORTAL: {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, CHAT_IFRAME, m.Option(ORIGIN)+S(m.Option(mdb.NAME))+C(PORTAL), arg...)
				kit.If(!kit.HasPrefixList(arg, ctx.RUN), func() { m.Push("title", m.Option(mdb.NAME)) })
			}},
		}, ctx.ConfAction(ctx.TOOLS, DREAM)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmd(SPIDE, arg, kit.Dict(ice.MSG_FIELDS, "time,client.type,client.name,client.origin")).Table(func(value ice.Maps) {
					if value[CLIENT_TYPE] == nfs.REPOS && value[CLIENT_NAME] != ice.OPS {
						m.Push(mdb.NAME, value[CLIENT_NAME])
					}
				})
				m.Action(html.FILTER, mdb.CREATE).Display("")
				ctx.Toolkit(m)
			} else {
				origin := SpideOrigin(m, arg[0])
				m.SetAppend().Spawn().SplitIndex(m.Cmdx(SPIDE, arg[0], C(DREAM))).Table(func(value ice.Maps) {
					if value[mdb.TYPE] != WORKER {
						return
					}
					m.Push("", value, kit.Split("time,name,icon,repos,binary,module,version"))
					m.Push(mdb.TEXT, kit.JoinLine(value[nfs.REPOS], value[nfs.BINARY]))
					if m.Push(ORIGIN, origin); !nfs.Exists(m, path.Join(ice.USR_LOCAL_WORK, value[mdb.NAME])) {
						m.PushButton(INSTALL, PORTAL)
					} else {
						m.PushButton(OPEN, PORTAL)
					}
				})
				m.StatusTimeCount(ORIGIN, origin)
			}
		}},
	})
}
