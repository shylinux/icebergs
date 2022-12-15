package lark

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _employee_info(m *ice.Message, appid string, arg ...string) {
	kit.For(arg, func(id string) {
		_, data := _lark_get(m, appid, "/open-apis/contact/v1/user/batch_get", "open_ids", id)
		kit.For(kit.Value(data, "data.user_infos"), func(value ice.Map) { m.PushDetail(value) })
	})
}
func _employee_openid(m *ice.Message, appid string, arg ...string) {
	args := []string{}
	for i := 0; i < len(arg); i++ {
		args = append(args, kit.Select("mobiles", "emails", strings.Contains(arg[i], "@")), arg[i])
	}
	_lark_get(m, appid, "/open-apis/user/v1/batch_get_id", args)
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
			} else if strings.HasPrefix(arg[1], "ou_") {
				_employee_info(m, arg[0], arg[1:]...)
			} else {
				_employee_openid(m, arg[0], arg[1:]...)
			}
		}},
	})
}
