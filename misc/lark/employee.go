package lark

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _employee_info(m *ice.Message, appid string, arg ...string) {
	for _, id := range arg {
		_, data := _lark_get(m, appid, "/open-apis/contact/v1/user/batch_get", "open_ids", id)
		kit.Fetch(kit.Value(data, "data.user_infos"), func(index int, value ice.Map) {
			m.Push(mdb.DETAIL, value)
		})
	}
}
func _employee_openid(m *ice.Message, appid string, arg ...string) {
	us := []string{}
	for i := 0; i < len(arg); i++ {
		us = append(us, kit.Select("mobiles", "emails", strings.Contains(arg[i], "@")), arg[i])
	}

	_lark_get(m, appid, "/open-apis/user/v1/batch_get_id", us)
	for i := 0; i < len(arg); i++ {
		m.Echo(m.Append(kit.Keys("data.mobile_users", arg[i], "0.open_id")))
	}
}

const EMPLOYEE = "employee"

func init() {
	Index.MergeCommands(ice.Commands{
		EMPLOYEE: {Name: "employee appid open_id|mobile|email auto", Help: "员工", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) < 2 {
				return
			}
			if strings.HasPrefix(arg[1], "ou_") {
				_employee_info(m, arg[0], arg[1:]...)
			} else {
				_employee_openid(m, arg[0], arg[1:]...)
			}
		}},
	})
}
