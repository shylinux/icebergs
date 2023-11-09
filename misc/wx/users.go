package wx

import (
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	OPENID = "openid"
)
const USERS = "users"

func init() {
	Index.MergeCommands(ice.Commands{
		USERS: {Name: "users access openid auto", Help: "用户", Meta: Meta(), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ACCESS).PushAction("").Option(ice.MSG_ACTION, "")
			} else if m.Options(ACCESS, arg[0]); len(arg) == 1 {
				res := SpideGet(m, USER_GET)
				kit.For(kit.Value(res, "data.openid"), func(value string) {
					res := SpideGet(m, USER_INFO, OPENID, value)
					m.Push(mdb.TIME, time.Unix(kit.Int64(kit.Value(res, "subscribe_time")), 0).Format(ice.MOD_TIME))
					m.Push("", res, []string{OPENID, "sex", aaa.USERNICK, aaa.LANGUAGE, aaa.PROVINCE, aaa.CITY})
				})
				m.StatusTimeCountTotal(kit.Value(res, mdb.TOTAL), mdb.NEXT, kit.Value(res, "next_openid"))
			} else {
				m.Push(ice.FIELDS_DETAIL, SpideGet(m, USER_INFO, OPENID, arg[1]))
			}
		}},
	})
}
