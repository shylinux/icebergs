package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _matrix_list(m *ice.Message, domain string) (server []string) {
	fields := kit.Split(mdb.Config(m, mdb.FIELD))
	button := []ice.Any{PORTAL, DESKTOP, ADMIN, VIMER, XTERM, OPEN, kit.Select(UPGRADE, COMPILE, domain == "")}
	value := kit.Dict(cli.ParseMake(m.Cmdx(Space(m, domain), cli.RUNTIME)))
	value[mdb.TYPE], value[mdb.ICONS] = SERVER, kit.Select(nfs.USR_ICONS_ICEBERGS, ice.SRC_MAIN_ICO, domain == "")
	value[DOMAIN] = kit.Select(ice.CONTEXTS, domain)
	m.PushRecord(value, fields...).PushButton(button...)
	m.Cmd(Space(m, domain), DREAM, ice.Maps{"space.timeout": "3s", "dream.simple": ice.TRUE}).Table(func(value ice.Maps) {
		switch value[mdb.TYPE] {
		case WORKER:
			value[DOMAIN] = kit.Select(ice.CONTEXTS, domain)
			m.PushRecord(value, fields...).PushButton(button...)
		case SERVER:
			server = append(server, kit.Keys(domain, value[mdb.NAME]))
		}
	})
	return
}
func _matrix_cmd(m *ice.Message, action string, arg ...string) {
	domain := kit.Keys(kit.Select("", m.Option(DOMAIN), m.Option(DOMAIN) != ice.CONTEXTS), m.Option(mdb.NAME))
	switch action {
	case PORTAL, DESKTOP, ADMIN, OPEN:
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

const MATRIX = "matrix"

func init() {
	Index.MergeCommands(ice.Commands{
		MATRIX: {Name: "matrix list", Help: "空间矩阵", Actions: ice.MergeActions(ice.Actions{
			INSTALL: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(Space(m, m.Option(DOMAIN)), DREAM, ctx.ACTION, mdb.CREATE, m.OptionSimple(mdb.NAME, mdb.ICONS, nfs.REPOS), nfs.BINARY, UserHost(m)+S(m.Option(mdb.NAME)))
				m.Cmd(Space(m, m.Option(DOMAIN)), DREAM, ctx.ACTION, cli.START, m.OptionSimple(mdb.NAME))
			}},
		}, ctx.ConfAction(mdb.FIELD, "time,domain,status,type,name,text,icons,repos,binary,module,version")), Hand: func(m *ice.Message, arg ...string) {
			if kit.HasPrefixList(arg, ctx.ACTION) {
				_matrix_cmd(m, arg[1], arg[2:]...)
				return
			}
			GoToast(m, "", func(toast func(name string, count, total int)) []string {
				kit.For(_matrix_list(m, ""), func(domain string, index int, total int) {
					toast(domain, index, total)
					_matrix_list(m, domain)
				})
				return nil
			}).Sort("name,domain", "str_r", "str_r").Display("")
		}},
	})
}
