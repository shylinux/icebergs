package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const OCEAN = "ocean"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		OCEAN: {Name: "ocean username auto insert invite", Help: "用户", Action: map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(aaa.USER, ice.OptionFields(aaa.USERNAME, aaa.USERNICK, aaa.USERZONE))
			}},
			mdb.INSERT: {Name: "insert username", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m, OCEAN), mdb.HASH, arg)
			}},
			mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, RIVER, _river_key(m, OCEAN), mdb.HASH, m.OptionSimple(aaa.USERNAME))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(arg), "time,username")
			m.Cmdy(mdb.SELECT, RIVER, _river_key(m, OCEAN), mdb.HASH, aaa.USERNAME, arg)
			m.Table(func(index int, value map[string]string, head []string) {
				msg := m.Cmd(aaa.USER, value[aaa.USERNAME])
				m.Push(aaa.USERNICK, msg.Append(aaa.USERNICK))
				m.PushImages(aaa.AVATAR, msg.Append(aaa.AVATAR), kit.Select("60", "240", m.FieldsIsDetail()))
			})
			m.PushAction(mdb.DELETE)
		}},
	}})
}