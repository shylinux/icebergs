package web

import (
	"net/url"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

const STORE = "store"

func init() {
	Index.MergeCommands(ice.Commands{
		STORE: {Name: "store refresh", Help: "商店", Icon: "App Store.png", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				AddPortalProduct(m, "商店", `
每个用户都可以将自己的空间列表，以系统商店的方式分享给其它用户。
同样的每个用户，也可以添加任意多个商店，直接将空间下载到本机使用。
`, 300.0)
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(SPIDE, mdb.INPUTS, arg) }},
			mdb.CREATE: {Name: "create origin* name icons", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPIDE, mdb.CREATE, m.OptionSimple("origin,name,icons"), mdb.TYPE, nfs.REPOS)
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPIDE, mdb.REMOVE, CLIENT_NAME, m.Option(mdb.NAME))
			}},
			INSTALL: {Name: "install name*", Hand: func(m *ice.Message, arg ...string) {
				if !kit.HasPrefixList(arg, ctx.RUN) {
					m.Cmdy(DREAM, mdb.CREATE, m.OptionSimple(mdb.NAME, nfs.REPOS, nfs.BINARY))
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
			"connect": {Help: "连接", Hand: func(m *ice.Message, arg ...string) {
				m.Options(m.Cmd(SPIDE, m.Option(mdb.NAME)).AppendSimple())
				m.Cmdy(SPIDE, mdb.DEV_REQUEST)
			}},
		}, ctx.ConfAction(CLIENT_TIMEOUT, cli.TIME_3s), DREAM), Hand: func(m *ice.Message, arg ...string) {
			if kit.HasPrefixList(arg, ctx.ACTION) {
				m.Cmdy(DREAM, arg)
			} else if m.Display("").DisplayCSS(""); len(arg) == 0 {
				list := []string{}
				m.Cmd(SPIDE, arg, kit.Dict(ice.MSG_FIELDS, "time,icons,client.type,client.name,client.origin")).Table(func(value ice.Maps) {
					kit.If(value[CLIENT_TYPE] == nfs.REPOS && value[CLIENT_NAME] != ice.SHY, func() {
						list = append(list, value[CLIENT_NAME])
						if value[CLIENT_NAME] == ice.OPS {
							value[CLIENT_ORIGIN] = UserHost(m)
						}
						m.Push(mdb.NAME, value[CLIENT_NAME]).Push(mdb.ICONS, value[mdb.ICONS]).Push(ORIGIN, value[CLIENT_ORIGIN])
					})
				})
				if ice.Info.NodeType == WORKER || !aaa.IsTechOrRoot(m) {
					m.Action()
				} else {
					m.PushAction("connect", mdb.REMOVE).Action(mdb.CREATE)
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
				stat := map[string]int{}
				list := m.Spawn(ice.Maps{ice.MSG_FIELDS: ""}).CmdMap(DREAM, mdb.NAME)
				m.SetAppend().Spawn().SplitIndex(m.Cmdx(SPIDE, arg[0], dream, kit.Dict(mdb.ConfigSimple(m, CLIENT_TIMEOUT)))).Table(func(value ice.Maps) {
					stat[value[mdb.TYPE]]++
					if value[nfs.BINARY] == "" {
						value[nfs.BINARY] = origin + S(value[mdb.NAME])
					}
					m.Push("", value, kit.Split("time,type,name,icons,repos,binary,module,version"))
					if value[mdb.TYPE] == SERVER {
						m.Push(mdb.TEXT, value[mdb.TEXT]).Push(ORIGIN, value[mdb.TEXT]).PushButton()
						return
					}
					m.Push(mdb.TEXT, value[nfs.REPOS]).Push(ORIGIN, origin)
					if _, ok := list[value[mdb.NAME]]; ok || arg[0] == ice.OPS {
						m.PushButton(PORTAL, INSTALL)
					} else if ice.Info.NodeType == WORKER || !aaa.IsTechOrRoot(m) {
						m.PushButton(PORTAL)
					} else {
						m.PushButton(PORTAL, INSTALL)
					}
				})
				m.RewriteAppend(func(value, key string, index int) string {
					value, _ = url.QueryUnescape(value)
					return value
				})
				m.StatusTimeCount(ORIGIN, origin, stat)
			}
		}},
	})
}
