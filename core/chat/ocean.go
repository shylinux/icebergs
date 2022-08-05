package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const OCEAN = "ocean"

func init() {
	Index.MergeCommands(ice.Commands{
		OCEAN: {Name: "ocean username auto insert invite", Help: "用户", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				gdb.Watch(m, RIVER_CREATE)
			}},
			RIVER_CREATE: {Name: "river.create river template", Help: "建群", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(OCEAN, mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(aaa.USER, ice.OptionFields(aaa.USERNAME, aaa.USERNICK, aaa.USERZONE))
			}},
			mdb.INSERT: {Name: "insert username", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.SHORT, aaa.USERNAME)
				m.Cmdy(mdb.INSERT, m.Prefix(RIVER), _river_key(m, OCEAN), mdb.HASH, arg)
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, m.Prefix(RIVER), _river_key(m, OCEAN), mdb.HASH, m.OptionSimple(aaa.USERNAME))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Fields(len(arg), "time,username")
			m.Cmdy(mdb.SELECT, RIVER, _river_key(m, OCEAN), mdb.HASH, aaa.USERNAME, arg)
			m.Tables(func(value ice.Maps) {
				msg := m.Cmd(aaa.USER, value[aaa.USERNAME])
				m.Push(aaa.USERNICK, msg.Append(aaa.USERNICK))
				m.PushImages(aaa.AVATAR, msg.Append(aaa.AVATAR), kit.Select("60", "240", m.FieldsIsDetail()))
			})
			m.PushAction(mdb.DELETE)
		}},
	})
}
