package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _order_show(m *ice.Message, text string, arg ...string) {
	_wiki_template(m.Options(mdb.LIST, kit.SplitLine(text)), "", "", text, arg...)
}

const ORDER = "order"

func init() {
	Index.MergeCommands(ice.Commands{
		ORDER: {Name: "order text", Help: "列表", Hand: func(m *ice.Message, arg ...string) {
			_order_show(m, arg[0], arg[1:]...)
		}},
	})
}
