package code

import (
	"strings"
	"path"
	
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _css_stat(m *ice.Message, block string, stats map[string]int) {
	msg := m.Spawn()
	for k, v := range stats {
		msg.Push("name", k)
		msg.Push("value", v)
		msg.Push("block", block)
	}
	msg.SortIntR("value")
	m.Copy(msg)
}
func _css_show(m *ice.Message, arg ...string) {
	block := ""
	stats_key := map[string]int{}
	stats_value := map[string]int{}
	m.Cmd(nfs.CAT, path.Join(arg[2], arg[1]), func(line string) {
		if line = strings.TrimSpace(line); line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			return
		}
		switch {
		case strings.HasSuffix(line, "{"):
			block = strings.TrimSuffix(line, "{")
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
				m.Push("name", list[0])
				m.Push("value", list[1])
				m.Push("block", ls[0])
				stats_key[list[0]]++
				stats_value[list[1]]++
			}
		default:
		}
	})
	_css_stat(m, "stats.key", stats_key)
	_css_stat(m, "stats.value", stats_value)
	m.StatusTimeCount()
}
func _css_exec(m *ice.Message, arg ...string) {
	if arg[2] == "usr/volcanos/" && strings.HasPrefix(arg[1], "plugin/local/") {
		key := "web."+strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(arg[1], "plugin/local/"), ".css"), ice.PS, ice.PT)
		ctx.ProcessCommand(m, kit.Select("can.plugin", key), kit.Simple())
		return
	}
}

const CSS = "css"

func init() {
	Index.MergeCommands(ice.Commands{
		CSS: {Name: "css path auto", Help: "样式表", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _css_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _css_exec(m, arg...) }},
		}, PlugAction(), LangAction())},
	})
}