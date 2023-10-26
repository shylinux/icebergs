package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const ORDER = "order"

func init() {
	Index.MergeCommands(ice.Commands{
		ORDER: {Name: "order hash auto", Help: "订单", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create zone* type name* text price* count* image*=4@img"},
		}), Hand: func(m *ice.Message, arg ...string) {

		}},
	})
}
