package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _table_run(m *ice.Message, arg ...string) {
	msg := m.Cmd(arg)
	list := [][]string{}
	msg.Table(func(index int, value ice.Maps, head []string) {
		if index == 0 {
			m.Optionv("head", head)
		}
		line := []string{}
		for _, h := range head {
			line = append(line, value[h])
		}
		list = append(list, line)
	})
	m.Optionv("list", list)
	_wiki_template(m, "", "")
}
func _table_show(m *ice.Message, text string, arg ...string) {
	head, list := []string{}, [][]string{}
	for i, line := range kit.SplitLine(text) {
		if line = strings.ReplaceAll(line, "%", "%%"); i == 0 {
			head = kit.SplitWord(line)
			continue
		}
		list = append(list, transList(kit.SplitWord(line), func(value string) string {
			if ls := kit.SplitWord(value); len(ls) > 1 {
				return kit.Format(`<span style="%s">%s</span>`, kit.JoinKV(":", ";", transArgKey(ls[1:])...), ls[0])
			}
			return value
		}))
	}
	m.Optionv("head", head)
	m.Optionv("list", list)
	_wiki_template(m, "", text, arg...)
}
func transList(arg []string, cb func(string) string) []string {
	for i, v := range arg {
		arg[i] = cb(v)
	}
	return arg
}
func transArgKey(arg []string) []string {
	for i := 0; i < len(arg)-1; i += 2 {
		switch arg[i] {
		case BG:
			arg[i] = "background-color"
		case FG:
			arg[i] = "color"
		}
	}
	return arg
}

const TABLE = "table"

func init() {
	Index.MergeCommands(ice.Commands{
		TABLE: {Name: "table text", Help: "表格", Actions: ice.MergeActions(ice.Actions{
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) { _table_run(m, arg...) }},
		}, WordAction(`<table {{.OptionTemplate}}>
<tr>{{range $i, $v := .Optionv "head"}}<th>{{$v}}</th>{{end}}</tr>
{{range $index, $value := .Optionv "list"}}<tr>{{range $i, $v := $value}}<td>{{$v}}</td>{{end}}</tr>{{end}}
</table>`)), Hand: func(m *ice.Message, arg ...string) { _table_show(m, arg[0], arg[1:]...) }},
	})
}
