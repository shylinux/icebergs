package web

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const STORE = "store"

func init() {
	Index.MergeCommands(ice.Commands{
		STORE: {Name: "store refresh", Help: "系统商店", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(SPIDE, mdb.INPUTS, arg) }},
			mdb.CREATE: {Name: "create name* origin* icons", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPIDE, mdb.CREATE, m.OptionSimple("name,origin,icons"), mdb.TYPE, nfs.REPOS)
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPIDE, mdb.REMOVE, CLIENT_NAME, m.Option(mdb.NAME))
			}},
			INSTALL: {Hand: func(m *ice.Message, arg ...string) {
				if !kit.HasPrefixList(arg, ctx.RUN) {
					if strings.HasPrefix(m.Option(mdb.ICON), nfs.REQUIRE) {
						m.Option(mdb.ICON, strings.TrimSuffix(strings.TrimPrefix(m.Option(mdb.ICON), nfs.REQUIRE), "?pod="+m.Option(mdb.NAME)))
					}
					m.OptionDefault(nfs.BINARY, m.Option(ORIGIN)+S(m.Option(mdb.NAME)))
					m.Cmdy(DREAM, mdb.CREATE, m.OptionSimple(mdb.NAME, mdb.ICON, nfs.REPOS, nfs.BINARY))
					m.Cmdy(DREAM, cli.START, m.OptionSimple(mdb.NAME))
				}
				ProcessIframe(m, m.Option(mdb.NAME), S(m.Option(mdb.NAME)), arg...)
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) {
				ProcessIframe(m, m.Option(mdb.NAME), S(m.Option(mdb.NAME)), arg...)
			}},
			PORTAL: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				ProcessIframe(m, m.Option(mdb.NAME), m.Option(ORIGIN)+S(m.Option(mdb.NAME))+C(PORTAL), arg...)
			}},
		}, ctx.ConfAction(CLIENT_TIMEOUT, cli.TIME_3s), DREAM), Hand: func(m *ice.Message, arg ...string) {
			if kit.HasPrefixList(arg, ctx.ACTION) {
				m.Cmdy(DREAM, arg)
				return
			}
			if m.Display(""); len(arg) == 0 {
				m.Cmd(SPIDE, arg, kit.Dict(ice.MSG_FIELDS, "time,icons,client.type,client.name,client.origin")).Table(func(value ice.Maps) {
					kit.If(value[CLIENT_TYPE] == nfs.REPOS, func() { m.Push(mdb.NAME, value[CLIENT_NAME]).Push(mdb.ICONS, value[mdb.ICONS]) })
				})
				if ice.Info.NodeType == WORKER || !aaa.IsTechOrRoot(m) {
					m.Action()
				} else {
					m.PushAction(mdb.REMOVE).Action(html.FILTER, mdb.CREATE)
				}
			} else {
				if arg[0] == ice.OPS && ice.Info.NodeType == SERVER {
					m.Cmdy(DREAM)
					return
				}
				defer ToastProcess(m, m.Trans("querying, please wait", "查询中，请稍候"))(m.Trans("query succsess", "查询成功"))
				origin := SpideOrigin(m, arg[0])
				kit.If(kit.IsIn(arg[0], ice.OPS, ice.DEV), func() { origin = tcp.PublishLocalhost(m, origin) })
				list := m.Spawn(ice.Maps{ice.MSG_FIELDS: ""}).CmdMap(DREAM, mdb.NAME)
				m.SetAppend().Spawn().SplitIndex(m.Cmdx(SPIDE, arg[0], C(DREAM), kit.Dict(mdb.ConfigSimple(m, CLIENT_TIMEOUT)))).Table(func(value ice.Maps) {
					if value[mdb.TYPE] != WORKER {
						return
					}
					m.Push("", value, kit.Split("time,name,icon,repos,binary,module,version"))
					m.Push(mdb.TEXT, value[nfs.REPOS]).Push(ORIGIN, origin)
					if _, ok := list[value[mdb.NAME]]; ok || arg[0] == ice.OPS {
						m.PushButton(OPEN, PORTAL)
					} else if ice.Info.NodeType == WORKER || !aaa.IsTechOrRoot(m) {
						m.PushButton(PORTAL)
					} else {
						m.PushButton(INSTALL, PORTAL)
					}
				})
				m.StatusTimeCount(ORIGIN, origin)
			}
		}},
	})
}
