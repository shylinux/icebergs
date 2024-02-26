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

func _matrix_list(m *ice.Message, domain, icon string, fields ...string) (server, icons []string) {
	value := kit.Dict(cli.ParseMake(m.Cmdx(Space(m, domain), cli.RUNTIME)))
	value[DOMAIN], value[mdb.TYPE], value[mdb.ICONS] = domain, SERVER, icon
	button := []ice.Any{PORTAL, DESKTOP, ADMIN, OPEN, UPGRADE, cli.RUNTIME, DREAM, WORD, STATUS, VIMER, XTERM}
	if domain == "" {
		button = []ice.Any{PORTAL, WORD, STATUS, VIMER, COMPILE, cli.RUNTIME, XTERM, DESKTOP, DREAM, ADMIN, OPEN}
	}
	m.PushRecord(value, fields...).PushButton(button...)
	button = []ice.Any{PORTAL, DESKTOP, ADMIN, OPEN, UPGRADE, cli.RUNTIME, WORD, STATUS, VIMER, XTERM}
	if domain == "" {
		button = []ice.Any{PORTAL, WORD, STATUS, VIMER, COMPILE, cli.RUNTIME, XTERM, DESKTOP, ADMIN, OPEN}
	}
	button = append(button, cli.STOP)
	m.Cmd(Space(m, domain), DREAM).Table(func(value ice.Maps) {
		switch value[mdb.TYPE] {
		case WORKER:
			value[DOMAIN] = domain
			kit.If(value[mdb.STATUS] == cli.STOP, func() { value[mdb.ICONS] = nfs.USR_ICONS_ICEBERGS })
			kit.If(value[mdb.STATUS] == cli.STOP, func() { button = []ice.Any{cli.START, mdb.REMOVE} })
			m.PushRecord(value, fields...).PushButton(button...)
		case SERVER, MASTER:
			server = append(server, kit.Keys(domain, value[mdb.NAME]))
			icons = append(icons, value[mdb.ICONS])
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
			title, link := kit.Keys(domain, kit.Select("", action, action != OPEN)), kit.Select("", S(domain), domain != "")+kit.Select("", C(action), action != OPEN)
			ProcessIframe(m, kit.Select(ice.CONTEXTS, title), kit.Select(nfs.PS, link), arg...).ProcessField(ctx.ACTION, action, ctx.RUN)
		}
	case OPEN:
		m.ProcessOpen(kit.Select(nfs.PS, S(domain), domain != ""))
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

const MATRIX = "matrix"

func init() {
	Index.MergeCommands(ice.Commands{
		MATRIX: {Name: "matrix refresh", Help: "矩阵", Meta: kit.Dict(
			ice.CTX_ICONS, kit.Dict(STATUS, "bi bi-git"), ice.CTX_TRANS, kit.Dict(STATUS, "源码"),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(DREAM, mdb.INPUTS, arg) }},
			mdb.CREATE: {Name: "create name*=hi icons repos binary template", Hand: func(m *ice.Message, arg ...string) { m.Cmd(DREAM, mdb.CREATE, arg) }},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) { _matrix_dream(m, nfs.TRASH); _matrix_dream(m, "") }},
			cli.START:  {Hand: func(m *ice.Message, arg ...string) { _matrix_dream(m, "") }},
			cli.STOP:   {Hand: func(m *ice.Message, arg ...string) { _matrix_dream(m, "") }},
			COMPILE: {Hand: func(m *ice.Message, arg ...string) {
				_matrix_cmd(m, "", cli.AMD64, cli.LINUX, ice.SRC_MAIN_GO).ProcessHold()
			}},
			UPGRADE: {Hand: func(m *ice.Message, arg ...string) { _matrix_cmd(m, "").ProcessRefresh().Sleep3s() }},
			INSTALL: {Hand: func(m *ice.Message, arg ...string) {
				if kit.IsIn(m.Cmdv(Space(m, m.Option(DOMAIN)), SPIDE, ice.DEV_IP, CLIENT_HOSTNAME), m.Cmd(tcp.HOST).Appendv(aaa.IP)...) {
					m.Option(nfs.BINARY, S(m.Option(mdb.NAME)))
				} else {
					m.OptionDefault(nfs.BINARY, UserHost(m)+S(m.Option(mdb.NAME)))
				}
				_matrix_dream(m, mdb.CREATE, kit.Simple(m.OptionSimple(mdb.ICONS, nfs.REPOS, nfs.BINARY))...)
			}},
		}, ctx.ConfAction(mdb.FIELD, "time,domain,status,type,name,text,icons,repos,binary,module,version", ctx.TOOLS, kit.Simple(STATUS, VERSION))), Hand: func(m *ice.Message, arg ...string) {
			if kit.HasPrefixList(arg, ctx.ACTION) {
				_matrix_action(m, arg[1], arg[2:]...)
				return
			}
			GoToast(m, func(toast func(name string, count, total int)) []string {
				field := kit.Split(mdb.Config(m, mdb.FIELD))
				m.Options("space.timeout", cli.TIME_300ms, "dream.simple", ice.TRUE)
				list, icons := _matrix_list(m, "", ice.SRC_MAIN_ICO, field...)
				kit.For(list, func(domain string, index int, total int) {
					toast(domain, index, total)
					_matrix_list(m, domain, icons[index], field...)
				})
				m.RewriteAppend(func(value, key string, index int) string {
					if key == mdb.ICONS && strings.HasPrefix(value, nfs.REQUIRE) && m.Appendv(DOMAIN)[index] != "" {
						value = kit.MergeURL(strings.Split(value, "?")[0], ice.POD, kit.Keys(m.Appendv(DOMAIN)[index], m.Appendv(mdb.NAME)[index]))
					}
					return value
				})
				m.Sort("type,status,name,domain", []string{MASTER, SERVER, WORKER, ""}, []string{cli.START, cli.STOP, ""}, "str_r", "str")
				m.StatusTimeCountStats(mdb.TYPE, mdb.STATUS)
				return nil
			}).Action(html.FILTER, mdb.CREATE).Display("")
			ctx.Toolkit(m)
		}},
	})
}
