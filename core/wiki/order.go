package wiki

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _order_show(m *ice.Message, name, text string, arg ...string) {
	m.Optionv("list", kit.Split(strings.TrimSpace(text), "\n"))
	_option(m, ORDER, name, text, arg...)
	m.RenderTemplate(m.Conf(ORDER, kit.Keym(kit.MDB_TEMPLATE)))
}

const ORDER = "order"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			ORDER: {Name: "order [name] `[item \n]...`", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_order_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		},
		Configs: map[string]*ice.Config{
			ORDER: {Name: ORDER, Help: "列表", Value: kit.Data(
				kit.MDB_TEMPLATE, `<ul {{.OptionTemplate}}>{{range $index, $value := .Optionv "list"}}<li>{{$value}}</li>{{end}}</ul>`,
			)},
		},
	})
}
