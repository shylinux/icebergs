package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const (
	OPENID = "openid"
)
const (
	SUBSCRIBE_TIME = "subscribe_time"

	REMARK = "remark"
	SEX    = "sex"
)
const USERS = "users"

func init() {
	Index.MergeCommands(ice.Commands{
		USERS: {Name: "users access openid auto", Help: "用户", Meta: Meta(), Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case TAGS:
					m.Cmdy(TAGS, m.Option(ACCESS)).CutTo(mdb.NAME, arg[0])
				}
			}},
			TAGS: {Name: "tags tags", Help: "标签", Hand: func(m *ice.Message, arg ...string) {
				list := map[string]string{}
				m.Cmd(TAGS, m.Option(ACCESS), func(value ice.Maps) { list[value[mdb.NAME]] = value[mdb.ID] })
				SpidePost(m, TAGS_MEMBERS_BATCHTAGGING, TAGID, m.Option(TAGID, list[m.Option(TAGS)]), "openid_list.0", m.Option(OPENID))
			}},
		}), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ACCESS).PushAction("").Option(ice.MSG_ACTION, "")
			} else if m.Options(ACCESS, arg[0]); len(arg) == 1 {
				_user_list(m, SpideGet(m, USER_GET)).PushAction(TAGS)
			} else {
				_user_info(m, arg[1])
			}
		}},
	})
}
func _user_list(m *ice.Message, res ice.Any) *ice.Message {
	list := map[string]string{}
	m.Cmd(TAGS, m.Option(ACCESS), func(value ice.Maps) { list[value[mdb.ID]] = value[mdb.NAME] })
	kit.For(kit.Value(res, kit.Keys(mdb.DATA, OPENID)), func(value string) {
		res := SpideGet(m, USER_INFO, OPENID, value)
		m.Push(mdb.TIME, kit.TimeUnix(kit.Value(res, SUBSCRIBE_TIME)))
		m.Push("", res, []string{OPENID, SEX, aaa.USERNICK, aaa.LANGUAGE, aaa.PROVINCE, aaa.CITY})
		m.Push(TAGS, kit.Join(kit.Simple(kit.Value(res, "tagid_list"), func(id string) string { return list[id] })))
		m.Push(REMARK, kit.Format(kit.Value(res, REMARK)))
	})
	return m.StatusTimeCount(mdb.NEXT, kit.Value(res, "next_openid"))
	return m.StatusTimeCountTotal(kit.Value(res, mdb.TOTAL), mdb.NEXT, kit.Value(res, "next_openid"))
}
func _user_info(m *ice.Message, openid string) *ice.Message {
	m.Push(ice.FIELDS_DETAIL, SpideGet(m, USER_INFO, OPENID, openid))
	m.RewriteAppend(func(value, key string, index int) string {
		kit.If(key == SUBSCRIBE_TIME, func() { value = kit.TimeUnix(value) })
		return value
	})
	return m
}
