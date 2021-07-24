package lark

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"
)

func _company_list(m *ice.Message, appid string) {
	_, data := _lark_get(m, appid, "/open-apis/contact/v1/scope/get/")

	kit.Fetch(kit.Value(data, "data.authed_departments"), func(index int, ship_id string) {
		_, data := _lark_get(m, appid, "/open-apis/contact/v1/department/detail/batch_get", "department_ids", ship_id)
		kit.Fetch(kit.Value(data, "data.department_infos"), func(index int, value map[string]interface{}) {
			m.Push(SHIP_ID, ship_id)
			m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
			m.Push(kit.MDB_COUNT, value["member_count"])
			m.Push(CHAT_ID, value[CHAT_ID])
		})
	})
	m.Sort(kit.MDB_NAME)
}
func _company_members(m *ice.Message, appid string, ship_id string) {
	_, data := _lark_get(m, appid, "/open-apis/contact/v1/department/user/list",
		"department_id", ship_id, "page_size", "100", "fetch_child", "true")

	kit.Fetch(kit.Value(data, "data.user_list"), func(index int, value map[string]interface{}) {
		msg := m.Cmd(EMPLOYEE, appid, value[OPEN_ID])
		m.PushImages(aaa.AVATAR, msg.Append("avatar_72"))
		m.Push(aaa.GENDER, kit.Select("女", "男", msg.Append(aaa.GENDER) == "1"))
		m.Push(kit.MDB_NAME, msg.Append(kit.MDB_NAME))
		m.Push(kit.MDB_TEXT, msg.Append("description"))
		m.Push(OPEN_ID, msg.Append(OPEN_ID))
	})
	m.Sort(kit.MDB_NAME)
}

const COMPANY = "company"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		COMPANY: {Name: "company appid ship_id open_id text auto", Help: "组织", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			switch len(arg) {
			case 0: // 应用列表
				m.Cmdy(APP)

			case 1: // 组织列表
				_company_list(m, arg[0])

			case 2: // 员工列表
				_company_members(m, arg[0], arg[1])

			case 3: // 员工详情
				m.Cmdy(EMPLOYEE, arg[0], arg[2])

			default: // 员工通知
				m.Cmdy(SEND, arg[0], OPEN_ID, arg[2], arg[3:])
			}
		}},
	}})
}
