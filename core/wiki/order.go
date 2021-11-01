package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _order_show(m *ice.Message, text string, arg ...string) {
	m.Optionv(kit.MDB_LIST, kit.Split(strings.TrimSpace(text), ice.NL))
	_wiki_template(m, ORDER, "", text, arg...)
}

const ORDER = "order"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		ORDER: {Name: "order `[item\n]...`", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_order_show(m, arg[0], arg[1:]...)
		}},
	}, Configs: map[string]*ice.Config{
		ORDER: {Name: ORDER, Help: "列表", Value: kit.Data(
			kit.MDB_TEMPLATE, `<ul {{.OptionTemplate}}>{{range $index, $value := .Optionv "list"}}<li>{{$value}}</li>{{end}}</ul>`,
		)},
	}})
}
