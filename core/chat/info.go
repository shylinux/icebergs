package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const INFO = "info"

func init() {
	Index.MergeCommands(ice.Commands{
		INFO: {Name: "info auto", Help: "信息", Actions: ice.Actions{
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, RIVER, "", mdb.HASH, mdb.HASH, m.Option(ice.MSG_RIVER), arg)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.OptionFields(mdb.DETAIL)
			m.Cmdy(mdb.SELECT, RIVER, "", mdb.HASH, mdb.HASH, m.Option(ice.MSG_RIVER))
		}},
	})
}
