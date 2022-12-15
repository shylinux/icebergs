package lark

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _group_list(m *ice.Message, appid string) {
	_, data := _lark_get(m, appid, "/open-apis/chat/v4/list")
	kit.For(kit.Value(data, "data.groups"), func(index int, value ice.Map) {
		m.Push(CHAT_ID, value[CHAT_ID])
		m.PushImages(aaa.AVATAR, kit.Format(value[aaa.AVATAR]))
		m.Push(mdb.NAME, value[mdb.NAME])
		m.Push(mdb.TEXT, value["description"])
		m.Push(OPEN_ID, value["owner_open_id"])
	})
	m.Sort(mdb.NAME)
}
func _group_members(m *ice.Message, appid string, chat_id string) {
	_, data := _lark_get(m, appid, "/open-apis/chat/v4/info", "chat_id", chat_id)
	kit.For(kit.Value(data, "data.members"), func(index int, value ice.Map) {
		msg := m.Cmd(EMPLOYEE, appid, value[OPEN_ID])
		m.PushImages(aaa.AVATAR, msg.Append("avatar_72"))
		m.Push(aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"))
		m.Push(mdb.NAME, msg.Append(mdb.NAME))
		m.Push(mdb.TEXT, msg.Append("description"))
		m.Push(OPEN_ID, msg.Append(OPEN_ID))
	})
	m.Sort(mdb.NAME)
}

const GROUP = "group"

func init() {
	Index.MergeCommands(ice.Commands{
		GROUP: {Name: "group appid chat_id open_id text auto", Help: "群组", Hand: func(m *ice.Message, arg ...string) {
			kit.Switch(len(arg),
				0, func() { m.Cmdy(APP) },
				1, func() { _group_list(m, arg[0]) },
				2, func() { _group_members(m, arg[0], arg[1]) },
				3, func() { m.Cmdy(EMPLOYEE, arg[0], arg[2]) },
				func() { m.Cmdy(SEND, arg[0], OPEN_ID, arg[2], arg[3:]) },
			)
		}},
	})
}
