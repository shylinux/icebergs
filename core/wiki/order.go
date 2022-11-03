package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _order_show(m *ice.Message, text string, arg ...string) {
	m.Optionv(mdb.LIST, kit.Split(strings.TrimSpace(text), ice.NL))
	_wiki_template(m, "", text, arg...)
}

const ORDER = "order"

func init() {
	Index.MergeCommands(ice.Commands{
		ORDER: {Name: "order text", Help: "列表", Actions: WordAction(
			`<ul {{.OptionTemplate}}>{{range $index, $value := .Optionv "list"}}<li>{{$value}}</li>{{end}}</ul>`,
		), Hand: func(m *ice.Message, arg ...string) { _order_show(m, arg[0], arg[1:]...) }},
	})
}
