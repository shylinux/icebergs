package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _order_show(m *ice.Message, text string, arg ...string) {
	m.Optionv(mdb.LIST, kit.Split(strings.TrimSpace(text), ice.NL))
	_wiki_template(m, ORDER, "", text, arg...)
}

const ORDER = "order"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		ORDER: {Name: "order `[item\n]...`", Help: "列表", Hand: func(m *ice.Message, arg ...string) {
			_order_show(m, arg[0], arg[1:]...)
		}},
	}, Configs: ice.Configs{
		ORDER: {Name: ORDER, Help: "列表", Value: kit.Data(
			nfs.TEMPLATE, `<ul {{.OptionTemplate}}>{{range $index, $value := .Optionv "list"}}<li>{{$value}}</li>{{end}}</ul>`,
		)},
	}})
}
