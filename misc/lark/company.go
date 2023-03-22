package lark

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _company_list(m *ice.Message, appid string) {
	_, data := _lark_get(m, appid, "/open-apis/contact/v1/scope/get/")
	kit.For(kit.Value(data, "data.authed_departments"), func(index int, ship_id string) {
		_, data := _lark_get(m, appid, "/open-apis/contact/v1/department/detail/batch_get", "department_ids", ship_id)
		kit.For(kit.Value(data, "data.department_infos"), func(index int, value ice.Map) {
			m.Push(SHIP_ID, ship_id)
			m.Push(mdb.NAME, value[mdb.NAME])
			m.Push(mdb.COUNT, value["member_count"])
			m.Push(CHAT_ID, value[CHAT_ID])
		})
	})
	m.Sort(mdb.NAME)
}
func _company_members(m *ice.Message, appid string, ship_id string) {
	_, data := _lark_get(m, appid, "/open-apis/contact/v1/department/user/list", "department_id", ship_id, "page_size", "100", "fetch_child", ice.TRUE)
	kit.For(kit.Value(data, "data.user_list"), func(index int, value ice.Map) {
		msg := m.Cmd(EMPLOYEE, appid, value[OPEN_ID])
		m.PushImages(aaa.AVATAR, msg.Append("avatar_72"))
		m.Push(aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"))
		m.Push(mdb.NAME, msg.Append(mdb.NAME))
		m.Push(mdb.TEXT, msg.Append("description"))
		m.Push(OPEN_ID, msg.Append(OPEN_ID))
	})
	m.Sort(mdb.NAME)
}

const COMPANY = "company"

func init() {
	Index.MergeCommands(ice.Commands{
		COMPANY: {Name: "company appid ship_id open_id text auto", Help: "组织", Hand: func(m *ice.Message, arg ...string) {
			kit.Switch(len(arg),
				0, func() { m.Cmdy(APP) },
				1, func() { _company_list(m, arg[0]) },
				2, func() { _company_members(m, arg[0], arg[1]) },
				3, func() { m.Cmdy(EMPLOYEE, arg[0], arg[2]) },
				func() { m.Cmdy(SEND, arg[0], OPEN_ID, arg[2], arg[3:]) },
			)
		}},
	})
}
