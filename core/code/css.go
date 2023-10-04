package code

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _css_stat(m *ice.Message, zone string, stats map[string]int) {
	msg := m.Spawn()
	kit.For(stats, func(k string, v int) { msg.Push(mdb.NAME, k).Push(mdb.VALUE, v).Push(mdb.ZONE, zone) })
	m.Copy(msg.SortIntR(mdb.VALUE))
}
func _css_show(m *ice.Message, arg ...string) {
	zone, stats_key, stats_value := "", map[string]int{}, map[string]int{}
	m.Cmd(nfs.CAT, path.Join(arg[2], arg[1]), func(line string) {
		if line = strings.TrimSpace(line); line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			return
		}
		switch {
		case strings.HasSuffix(line, "{"):
			zone = strings.TrimSuffix(line, "{")
		case strings.HasSuffix(line, "}"):
			if line == "}" {
				break
			}
			ls := strings.Split(strings.TrimSuffix(line, "}"), "{")
			for _, l := range strings.Split(ls[1], ";") {
				list := strings.Split(l, ":")
				if len(list) < 2 {
					continue
				}
				m.Push(mdb.NAME, list[0])
				m.Push(mdb.VALUE, list[1])
				m.Push(mdb.ZONE, ls[0])
				stats_key[list[0]]++
				stats_value[list[1]]++
			}
		default:
			list := kit.Split(line, "", ":;")
			m.Push(mdb.NAME, list[0])
			m.Push(mdb.VALUE, list[2])
			m.Push(mdb.ZONE, zone)
		}
	})
	_css_stat(m, "stats.key", stats_key)
	_css_stat(m, "stats.value", stats_value)
	m.StatusTimeCount()
}
func _css_exec(m *ice.Message, arg ...string) {
	if arg[2] == ice.USR_VOLCANOS {
		if strings.HasPrefix(arg[1], ice.PLUGIN_LOCAL) {
			ctx.ProcessField(m, kit.Select(ice.CAN_PLUGIN, "web."+strings.Replace(kit.TrimExt(strings.TrimPrefix(arg[1], ice.PLUGIN_LOCAL), JS), nfs.PS, nfs.PT, -1)), kit.Simple())
		}
	} else {
		ctx.ProcessField(m, kit.Select(ice.CAN_PLUGIN, ctx.GetFileCmd(kit.ExtChange(path.Join(arg[2], arg[1]), GO))), kit.Simple())
		m.Push(ctx.STYLE, require(arg[2], arg[1]))
	}
}

const CSS = "css"

func init() {
	Index.MergeCommands(ice.Commands{
		CSS: {Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _css_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _css_exec(m, arg...) }},
			TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(kit.Format(nfs.Template(m, "demo.css"), kit.Select(mdb.PLUGIN, ctx.GetFileCmd(kit.ExtChange(path.Join(arg[2], arg[1]), GO)))))
			}},
		}, PlugAction())},
	})
}
