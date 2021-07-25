package wiki

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _table_show(m *ice.Message, name, text string, arg ...string) {
	head, list := []string{}, [][]string{}
	for i, v := range kit.Split(strings.TrimSpace(text), "\n") {
		if v = strings.ReplaceAll(v, "%", "%%"); i == 0 {
			head = kit.Split(v)
		} else {
			line := kit.Split(v)
			for i, v := range line {
				if ls := kit.Split(v); len(ls) > 1 {
					style := []string{}
					for i := 1; i < len(ls)-1; i += 2 {
						switch ls[i] {
						case "bg":
							ls[i] = "background-color"
						case "fg":
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

	_option(m, TABLE, name, text, arg...)
	m.RenderTemplate(m.Conf(TABLE, kit.Keym(kit.MDB_TEMPLATE)))
}

const TABLE = "table"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			TABLE: {Name: "table [name] `[item item\n]...`", Help: "表格", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_table_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		},
		Configs: map[string]*ice.Config{
			TABLE: {Name: TABLE, Help: "表格", Value: kit.Data(
				kit.MDB_TEMPLATE, `<table class="story"
{{range $k, $v := .Optionv "extra"}}data-{{$k}}='{{$v}}'{{end}}
data-type="{{.Option "type"}}" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">
<tr>{{range $i, $v := .Optionv "head"}}<th>{{$v}}</th>{{end}}</tr>
{{range $index, $value := .Optionv "list"}}
<tr>{{range $i, $v := $value}}<td>{{$v}}</td>{{end}}</tr>
{{end}}
</table>`,
			)},
		},
	})
}
