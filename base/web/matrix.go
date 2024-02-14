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

func _matrix_list(m *ice.Message, domain string, fields ...string) (server []string) {
	value := kit.Dict(cli.ParseMake(m.Cmdx(Space(m, domain), cli.RUNTIME)))
	value[DOMAIN], value[mdb.TYPE], value[mdb.ICONS] = domain, SERVER, kit.Select(nfs.USR_ICONS_ICEBERGS, ice.SRC_MAIN_ICO, domain == "")
	button := []ice.Any{PORTAL, ADMIN, DESKTOP, XTERM, UPGRADE, cli.RUNTIME, WORD, STATUS, VIMER, OPEN}
	if domain == "" {
		button = []ice.Any{PORTAL, WORD, STATUS, VIMER, COMPILE, cli.RUNTIME, XTERM, ADMIN, DESKTOP, OPEN}
	}
	m.PushRecord(value, fields...).PushButton(button...)
	button = append(button, cli.STOP)
	m.Cmd(Space(m, domain), DREAM).Table(func(value ice.Maps) {
		switch value[mdb.TYPE] {
		case WORKER:
			value[DOMAIN] = domain
			kit.If(value[mdb.STATUS] == cli.STOP, func() { value[mdb.ICONS] = nfs.USR_ICONS_ICEBERGS })
			kit.If(value[mdb.STATUS] == cli.STOP, func() { button = []ice.Any{cli.START, mdb.REMOVE} })
			m.PushRecord(value, fields...).PushButton(button...)
		case SERVER:
			server = append(server, kit.Keys(domain, value[mdb.NAME]))
		}
	})
	return
}
func _matrix_action(m *ice.Message, action string, arg ...string) {
	switch domain := kit.Keys(m.Option(DOMAIN), m.Option(mdb.NAME)); action {
	case PORTAL, ADMIN, DESKTOP, OPEN:
		if kit.HasPrefixList(arg, ctx.RUN) {
			ProcessIframe(m, "", "", arg...)
		} else {
			title, link := kit.Keys(domain, action), m.MergePodCmd(domain, action)
			kit.If(action == OPEN, func() { title, link = domain, m.MergePod(domain) })
			ProcessIframe(m, title, link, arg...).ProcessField(ctx.ACTION, action, ctx.RUN)
		}
	default:
		if kit.HasPrefixList(arg, ctx.RUN) {
			ctx.ProcessFloat(m, action, arg, arg...)
		} else {
			m.Option(ice.POD, domain)
			kit.If(action == XTERM, func() { arg = []string{cli.SH} })
			ctx.ProcessFloat(m, action, arg, arg...).ProcessField(ctx.ACTION, action, ctx.RUN)
		}
	}
}
func _matrix_dream(m *ice.Message, action string, arg ...string) {
	m.Cmd(Space(m, m.Option(DOMAIN)), DREAM, kit.Select(m.ActionKey(), action), m.OptionSimple(mdb.NAME), arg)
}

const MATRIX = "matrix"

func init() {
	Index.MergeCommands(ice.Commands{
		MATRIX: {Name: "matrix refresh", Help: "矩阵", Meta: kit.Dict(
			ice.CTX_ICONS, kit.Dict(STATUS, "bi bi-git"),
			ice.CTX_TRANS, kit.Dict(STATUS, "源码"),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(DREAM, mdb.INPUTS, arg) }},
			mdb.CREATE: {Name: "create name*=hi icons repos binary template", Hand: func(m *ice.Message, arg ...string) { m.Cmd(DREAM, mdb.CREATE, arg) }},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) { _matrix_dream(m, nfs.TRASH); _matrix_dream(m, "") }},
			cli.START:  {Hand: func(m *ice.Message, arg ...string) { _matrix_dream(m, "") }},
			cli.STOP:   {Hand: func(m *ice.Message, arg ...string) { _matrix_dream(m, "") }},
			UPGRADE: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(Space(m, kit.Keys(m.Option(DOMAIN), m.Option(mdb.NAME))), UPGRADE).Sleep3s()
			}},
			INSTALL: {Hand: func(m *ice.Message, arg ...string) {
				m.OptionDefault(nfs.BINARY, UserHost(m)+S(m.Option(mdb.NAME)))
				if kit.IsIn(m.Cmdv(Space(m, m.Option(DOMAIN)), SPIDE, ice.DEV_IP, CLIENT_HOSTNAME), m.Cmd(tcp.HOST).Appendv(aaa.IP)...) {
					m.Option(nfs.BINARY, S(m.Option(mdb.NAME)))
				}
				_matrix_dream(m, mdb.CREATE, kit.Simple(m.OptionSimple(mdb.ICONS, nfs.REPOS, nfs.BINARY))...)
				_matrix_dream(m, cli.START)
			}},
		}, ctx.ConfAction(mdb.FIELD, "time,domain,status,type,name,text,icons,repos,binary,module,version")), Hand: func(m *ice.Message, arg ...string) {
			if kit.HasPrefixList(arg, ctx.ACTION) {
				_matrix_action(m, arg[1], arg[2:]...)
				return
			}
			GoToast(m, "", func(toast func(name string, count, total int)) []string {
				fields := kit.Split(mdb.Config(m, mdb.FIELD))
				m.Options("space.timeout", cli.TIME_300ms, "dream.simple", ice.TRUE)
				kit.For(_matrix_list(m, "", fields...), func(domain string, index int, total int) {
					toast(domain, index, total)
					_matrix_list(m, domain, fields...)
				})
				stat := map[string]int{}
				m.Table(func(value ice.Maps) { stat[value[mdb.TYPE]]++; stat[value[mdb.STATUS]]++ }).StatusTimeCount(stat)
				m.RewriteAppend(func(value, key string, index int) string {
					if key == mdb.ICONS && strings.HasPrefix(value, nfs.REQUIRE) && m.Appendv(DOMAIN)[index] != "" {
						value = kit.MergeURL(strings.Split(value, "?")[0], ice.POD, kit.Keys(m.Appendv(DOMAIN)[index], m.Appendv(mdb.NAME)[index]))
					}
					return value
				}).Sort("type,status,name,domain", []string{SERVER, WORKER, ""}, []string{cli.START, cli.STOP, ""}, "str_r", "str_r")
				return nil
			}).Action(html.FILTER, mdb.CREATE).Display("")
		}},
	})
}
