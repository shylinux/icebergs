package lark

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	kit "shylinux.com/x/toolkits"
)

func _group_list(m *ice.Message, appid string) {
	_, data := _lark_get(m, appid, "/open-apis/chat/v4/list")
	kit.Fetch(kit.Value(data, "data.groups"), func(index int, value map[string]interface{}) {
		m.Push(CHAT_ID, value[CHAT_ID])
		m.PushImages(aaa.AVATAR, kit.Format(value[aaa.AVATAR]), "72")
		m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
		m.Push(kit.MDB_TEXT, value["description"])
		m.Push(OPEN_ID, value["owner_open_id"])
	})
	m.Sort(kit.MDB_NAME)
}
func _group_members(m *ice.Message, appid string, chat_id string) {
	_, data := _lark_get(m, appid, "/open-apis/chat/v4/info", "chat_id", chat_id)
	kit.Fetch(kit.Value(data, "data.members"), func(index int, value map[string]interface{}) {
		msg := m.Cmd(EMPLOYEE, appid, value[OPEN_ID])
		m.PushImages(aaa.AVATAR, msg.Append("avatar_72"))
		m.Push(aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"))
		m.Push(kit.MDB_NAME, msg.Append(kit.MDB_NAME))
		m.Push(kit.MDB_TEXT, msg.Append("description"))
		m.Push(OPEN_ID, msg.Append(OPEN_ID))
	})
	m.Sort(kit.MDB_NAME)
}

const GROUP = "group"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		GROUP: {Name: "group appid chat_id open_id text auto", Help: "群组", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			switch len(arg) {
			case 0: // 应用列表
				m.Cmdy(APP)

			case 1: // 群组列表
				_group_list(m, arg[0])

			case 2: // 组员列表
				_group_members(m, arg[0], arg[1])

			case 3: // 组员详情
				m.Cmdy(EMPLOYEE, arg[0], arg[2])

			default: // 组员通知
				m.Cmdy(SEND, arg[0], OPEN_ID, arg[2], arg[3:])
			}
		}},
	}})
}