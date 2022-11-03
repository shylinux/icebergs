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
	for i, v := range kit.SplitLine(text) {
		if v = strings.ReplaceAll(v, "%", "%%"); i == 0 {
			head = kit.SplitWord(v)
		} else {
			line := kit.SplitWord(v)
			for i, v := range line {
				if ls := kit.SplitWord(v); len(ls) > 1 {
					style := []string{}
					for i := 1; i < len(ls)-1; i += 2 {
						switch ls[i] {
						case BG:
							ls[i] = "background-color"
						case FG:
							ls[i] = "color"
						}
						style = append(style, ls[i]+":"+ls[i+1])
					}
					line[i] = kit.Format(`<span style="%s">%s</span>`, strings.Join(style, ";"), ls[0])
				}
			}
			list = append(list, line)
		}
	}
	m.Optionv("head", head)
	m.Optionv("list", list)
	_wiki_template(m, "", text, arg...)
}

const TABLE = "table"

func init() {
	Index.MergeCommands(ice.Commands{
		TABLE: {Name: "table text", Help: "表格", Actions: ice.MergeActions(ice.Actions{
			ice.RUN: {Name: "run", Hand: func(m *ice.Message, arg ...string) { _table_run(m, arg...) }},
		}, WordAction(`<table {{.OptionTemplate}}>
<tr>{{range $i, $v := .Optionv "head"}}<th>{{$v}}</th>{{end}}</tr>
{{range $index, $value := .Optionv "list"}}<tr>{{range $i, $v := $value}}<td>{{$v}}</td>{{end}}</tr>{{end}}
</table>`)), Hand: func(m *ice.Message, arg ...string) { _table_show(m, arg[0], arg[1:]...) }},
	})
}
