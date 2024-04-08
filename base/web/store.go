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
		STORE: {Name: "store refresh", Help: "商店", Icon: "App Store.png", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(SPIDE, mdb.INPUTS, arg) }},
			mdb.CREATE: {Name: "create origin* name icons", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPIDE, mdb.CREATE, m.OptionSimple("origin,name,icons"), mdb.TYPE, nfs.REPOS)
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
			PORTAL: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				ProcessIframe(m, m.Option(mdb.NAME), m.Option(ORIGIN)+S(m.Option(mdb.NAME))+C(m.ActionKey()), arg...)
			}},
			DESKTOP: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				ProcessIframe(m, kit.Keys(m.Option(mdb.NAME), m.ActionKey()), S(m.Option(mdb.NAME))+C(m.ActionKey()), arg...)
			}},
			ADMIN: {Role: aaa.VOID, Hand: func(m *ice.Message, arg ...string) {
				ProcessIframe(m, kit.Keys(m.Option(mdb.NAME), m.ActionKey()), S(m.Option(mdb.NAME))+C(m.ActionKey()), arg...)
			}},
			OPEN: {Hand: func(m *ice.Message, arg ...string) { m.ProcessOpen(S(m.Option(mdb.NAME))) }},
		}, ctx.ConfAction(CLIENT_TIMEOUT, cli.TIME_3s), DREAM), Hand: func(m *ice.Message, arg ...string) {
			if kit.HasPrefixList(arg, ctx.ACTION) {
				m.Cmdy(DREAM, arg)
				return
			}
			if m.Display(""); len(arg) == 0 {
				list := []string{}
				m.Cmd(SPIDE, arg, kit.Dict(ice.MSG_FIELDS, "time,icons,client.type,client.name,client.origin")).Table(func(value ice.Maps) {
					kit.If(value[CLIENT_TYPE] == nfs.REPOS, func() {
						list = append(list, value[CLIENT_NAME])
						m.Push(mdb.NAME, value[CLIENT_NAME]).Push(mdb.ICONS, value[mdb.ICONS]).Push(ORIGIN, value[CLIENT_ORIGIN])
					})
				})
				if ice.Info.NodeType == WORKER || !aaa.IsTechOrRoot(m) {
					m.Action()
				} else {
					m.PushAction(mdb.REMOVE).Action(html.FILTER, mdb.CREATE)
				}
			} else {
				defer ToastProcess(m, ice.LIST, arg[0])()
				if arg[0] == ice.OPS && ice.Info.NodeType == SERVER {
					m.Cmdy(DREAM)
					return
				}
				dream := C(DREAM)
				origin := SpideOrigin(m, arg[0])
				kit.If(origin == "", func() { arg[0], origin, dream = ice.DEV, arg[0], arg[0]+dream })
				if kit.IsIn(kit.ParseURL(origin).Hostname(), append(m.Cmds(tcp.HOST).Appendv(aaa.IP), tcp.LOCALHOST)...) {
					origin = m.Option(ice.MSG_USERHOST)
				} else {
					origin = tcp.PublishLocalhost(m, origin)
				}
				list := m.Spawn(ice.Maps{ice.MSG_FIELDS: ""}).CmdMap(DREAM, mdb.NAME)
				stat := map[string]int{}
				m.SetAppend().Spawn().SplitIndex(m.Cmdx(SPIDE, arg[0], dream, kit.Dict(mdb.ConfigSimple(m, CLIENT_TIMEOUT)))).Table(func(value ice.Maps) {
					stat[value[mdb.TYPE]]++
					m.Push("", value, kit.Split("time,type,name,icons,repos,binary,module,version"))
					if value[mdb.TYPE] == SERVER {
						m.Push(mdb.TEXT, value[mdb.TEXT]).Push(ORIGIN, value[mdb.TEXT])
						m.PushButton()
						return
					}
					m.Push(mdb.TEXT, value[nfs.REPOS]).Push(ORIGIN, origin)
					if _, ok := list[value[mdb.NAME]]; ok || arg[0] == ice.OPS {
						m.PushButton(PORTAL, DESKTOP, ADMIN, OPEN)
					} else if ice.Info.NodeType == WORKER || !aaa.IsTechOrRoot(m) {
						m.PushButton(PORTAL)
					} else {
						m.PushButton(PORTAL, INSTALL)
					}
				})
				m.StatusTimeCount(ORIGIN, origin, stat)
			}
		}},
	})
}
