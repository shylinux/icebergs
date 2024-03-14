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

func _matrix_list(m *ice.Message, domain, typ string, value ice.Maps, fields ...string) (server, icons, types []string) {
	value[DOMAIN], value[mdb.TYPE] = domain, typ
	istech, isdebug := typ == SERVER || kit.IsIn(value[aaa.ACCESS], aaa.TECH, aaa.ROOT), m.IsDebug()
	compile := kit.Select("", kit.Select(COMPILE, UPGRADE, typ == SERVER), istech)
	vimer := kit.Select("", VIMER, istech && isdebug)

	button := []ice.Any{PORTAL, DESKTOP, DREAM, ADMIN, OPEN, compile, MESSAGE}
	kit.If(istech, func() { button = append(button, WORD, STATUS) })
	kit.If(istech && isdebug, func() { button = append(button, vimer, cli.RUNTIME, XTERM) })
	m.PushRecord(value, fields...).PushButton(button...)

	button = []ice.Any{PORTAL, DESKTOP, MESSAGE, ADMIN, OPEN, compile}
	kit.If(istech, func() { button = append(button, WORD, STATUS) })
	kit.If(istech && isdebug, func() { button = append(button, vimer, cli.RUNTIME, XTERM, cli.STOP) })
	m.Cmd(Space(m, domain), DREAM).Table(func(value ice.Maps) {
		switch value[mdb.TYPE] {
		case WORKER:
			if value[aaa.ACCESS] == aaa.PRIVATE {
				break
			}
			value[DOMAIN] = domain
			kit.If(value[mdb.STATUS] == cli.STOP, func() { value[mdb.ICONS] = nfs.USR_ICONS_ICEBERGS })
			kit.If(value[mdb.STATUS] == cli.STOP && istech, func() { button = []ice.Any{cli.START, mdb.REMOVE} })
			m.PushRecord(value, fields...).PushButton(button...)
		case SERVER, ORIGIN:
			server = append(server, kit.Keys(domain, value[mdb.NAME]))
			icons = append(icons, value[mdb.ICONS])
			types = append(types, value[mdb.TYPE])
		}
	})
	return
}
func _matrix_action(m *ice.Message, action string, arg ...string) {
	switch domain := kit.Keys(m.Option(DOMAIN), m.Option(mdb.NAME)); action {
	case PORTAL, ADMIN:
		if kit.HasPrefixList(arg, ctx.RUN) {
			ProcessIframe(m, "", "", arg...)
		} else {
			title, link := kit.Keys(domain, action), kit.Select("", S(domain), domain != "")+C(action)
			if m.Option(mdb.TYPE) == ORIGIN {
				link = kit.MergeURL2(SpideOrigin(m, m.Option(DOMAIN)), C(action))
				if kit.IsIn(action, ADMIN) {
					m.ProcessOpen(link)
					break
				}
			}
			ProcessIframe(m, title, kit.Select(nfs.PS, link), arg...).ProcessField(ctx.ACTION, action, ctx.RUN)
		}
	case OPEN:
		link := kit.Select(nfs.PS, S(domain), domain != "")
		if m.Option(mdb.TYPE) == ORIGIN {
			link = SpideOrigin(m, m.Option(DOMAIN))
		} else if m.Option("server.type") == ORIGIN {
			link = kit.MergeURL2(SpideOrigin(m, m.Option(DOMAIN)), S(m.Option(mdb.NAME)))
		}
		m.ProcessOpen(link)
	default:
		if !kit.HasPrefixList(arg, ctx.RUN) {
			kit.If(action == XTERM, func() { arg = []string{cli.SH} })
			defer m.ProcessField(ctx.ACTION, action, ctx.RUN, domain, action)
		}
		ProcessPodCmd(m, domain, action, arg, arg...)
	}
}
func _matrix_dream(m *ice.Message, action string, arg ...string) {
	m.Cmd(Space(m, m.Option(DOMAIN)), DREAM, kit.Select(m.ActionKey(), action), m.OptionSimple(mdb.NAME), arg)
}
func _matrix_cmd(m *ice.Message, cmd string, arg ...string) *ice.Message {
	return m.Cmdy(Space(m, kit.Keys(m.Option(DOMAIN), m.Option(mdb.NAME))), kit.Select(m.ActionKey(), cmd), arg)
}

const (
	TARGET = "target"
)
const MATRIX = "matrix"

func init() {
	Index.MergeCommands(ice.Commands{
		MATRIX: {Name: "matrix refresh", Help: "矩阵", Icon: "Mission Control.png", Meta: kit.Dict(
			ice.CTX_ICONS, kit.Dict(STATUS, "bi bi-git"), ice.CTX_TRANS, kit.Dict(
				STATUS, "源码", html.INPUT, kit.Dict(MYSELF, "本机", ORIGIN, "主机"),
			),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(SPIDE, mdb.INPUTS, arg) }},
			mdb.CREATE: {Name: "create origin* name icons", Hand: func(m *ice.Message, arg ...string) {
				if u := kit.ParseURL(m.Option(ORIGIN)); m.Warn(u.Query().Get(TOKEN) == "", ice.ErrNotValid, TOKEN) {
					return
				} else {
					m.Option(TOKEN, u.Query().Get(TOKEN))
					m.Option(ORIGIN, kit.Format("%s://%s", u.Scheme, u.Host))
					m.OptionDefault(mdb.NAME, kit.Split(u.Hostname(), ".")[0])
					m.Cmd(SPIDE, mdb.CREATE, m.OptionSimple("name,origin,icons,token"), mdb.TYPE, nfs.REPOS)
					m.Cmd(SPIDE, DEV_CREATE_TOKEN, ice.Maps{CLIENT_NAME: m.Option(mdb.NAME)})
				}
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) { _matrix_dream(m, nfs.TRASH); _matrix_dream(m, "") }},
			cli.START:  {Hand: func(m *ice.Message, arg ...string) { _matrix_dream(m, "") }},
			cli.STOP:   {Hand: func(m *ice.Message, arg ...string) { _matrix_dream(m, "") }},
			UPGRADE: {Hand: func(m *ice.Message, arg ...string) {
				_matrix_cmd(m, "").Sleep3s()
				m.ProcessRefresh()
			}},
			INSTALL: {Hand: func(m *ice.Message, arg ...string) {
				if kit.IsIn(m.Cmdv(Space(m, m.Option(DOMAIN)), SPIDE, ice.DEV_IP, CLIENT_HOSTNAME), m.Cmd(tcp.HOST).Appendv(aaa.IP)...) {
					m.Option(nfs.BINARY, S(m.Option(mdb.NAME)))
				} else {
					m.OptionDefault(nfs.BINARY, UserHost(m)+S(m.Option(mdb.NAME)))
				}
				_matrix_dream(m, mdb.CREATE, kit.Simple(m.OptionSimple(mdb.ICONS, nfs.REPOS, nfs.BINARY))...)
				m.Cmd(SPACE, kit.Keys(m.Option(DOMAIN), m.Option(mdb.NAME)), MESSAGE, mdb.CREATE,
					mdb.TYPE, aaa.TECH, mdb.ICONS, nfs.USR_ICONS_VOLCANOS, TARGET, kit.Keys(nfs.FROM, m.Option(mdb.NAME)))
				m.Cmd(SPACE, m.Option(mdb.NAME), MESSAGE, mdb.CREATE,
					mdb.TYPE, aaa.TECH, mdb.ICONS, nfs.USR_ICONS_ICEBERGS, TARGET, kit.Keys(ice.OPS, m.Option(DOMAIN), m.Option(mdb.NAME)))
				StreamPushRefreshConfirm(m, m.Trans("refresh for new space ", "刷新列表查看新空间 ")+kit.Keys(m.Option(DOMAIN), m.Option(mdb.NAME)))
			}},
		}, ctx.ConfAction(
			mdb.FIELD, "time,domain,status,type,name,text,icons,repos,binary,module,version,access",
			ctx.TOOLS, kit.Simple(SPIDE, STATUS, VERSION), ONLINE, ice.TRUE,
		)), Hand: func(m *ice.Message, arg ...string) {
			if kit.HasPrefixList(arg, ctx.ACTION) {
				_matrix_action(m, arg[1], arg[2:]...)
				return
			}
			GoToast(m, func(toast func(name string, count, total int)) []string {
				field := kit.Split(mdb.Config(m, mdb.FIELD))
				space := m.CmdMap(SPACE, mdb.NAME)
				m.Options("space.timeout", cli.TIME_3s, "dream.simple", ice.TRUE)
				list, icons, types := _matrix_list(m, "", MYSELF, ice.Maps{
					mdb.TIME:    ice.Info.Make.Time,
					mdb.ICONS:   ice.SRC_MAIN_ICO,
					nfs.MODULE:  ice.Info.Make.Module,
					nfs.VERSION: ice.Info.Make.Versions(),
					aaa.ACCESS:  m.Option(ice.MSG_USERROLE),
				}, field...)
				kit.For(list, func(domain string, index int, total int) {
					toast(domain, index, total)
					_matrix_list(m, domain, types[index], ice.Maps{
						mdb.TIME:    space[domain][mdb.TIME],
						mdb.ICONS:   icons[index],
						nfs.MODULE:  space[domain][nfs.MODULE],
						nfs.VERSION: space[domain][nfs.VERSION],
						aaa.ACCESS:  kit.Format(kit.Value(space[domain], aaa.USERROLE)),
					}, field...)
				})
				m.RewriteAppend(func(value, key string, index int) string {
					if key == mdb.ICONS && strings.HasPrefix(value, nfs.REQUIRE) {
						if domain := m.Appendv(DOMAIN)[index]; domain != "" {
							value = kit.MergeURL2(space[domain][mdb.TEXT], value, ice.POD, m.Appendv(mdb.NAME)[index])
						}
					}
					return value
				})
				m.Action(html.FILTER, mdb.CREATE, UPGRADE).StatusTimeCountStats(mdb.TYPE, mdb.STATUS).Display("")
				m.Sort("type,status,name,domain", []string{MYSELF, SERVER, ORIGIN, WORKER, ""}, []string{cli.START, cli.STOP, ""}, ice.STR_R, ice.STR_R)
				ctx.Toolkit(m)
				return nil
			})
		}},
	})
}
