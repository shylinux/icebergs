package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _sess_create(m *ice.Message, username string, arg ...string) {
	if msg := m.Cmd(USER, username); msg.Length() > 0 {
		mdb.HashCreate(m, msg.AppendSimple(USERNICK, USERNAME, USERROLE), arg)
	} else {
		mdb.HashCreate(m, m.OptionSimple(USERNICK, USERNAME, USERROLE), arg)
	}
}
func _sess_check(m *ice.Message, sessid string) {
	if val := mdb.HashSelectDetails(m, sessid, func(value ice.Map) bool { return !m.WarnTimeNotValid(value[mdb.TIME], sessid) }); len(val) > 0 {
		SessAuth(m, val)
	}
}

const (
	UA = "ua"
	IP = "ip"
)
const (
	CHECK  = "check"
	LOGIN  = "login"
	LOGOUT = "logout"
)
const SESS = "sess"

func init() {
	Index.MergeCommands(ice.Commands{
		SESS: {Name: "sess hash auto prunes", Help: "会话", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create username*", Hand: func(m *ice.Message, arg ...string) {
				_sess_create(m, m.Option(USERNAME), UA, m.Option(ice.MSG_USERUA), IP, m.Option(ice.MSG_USERIP))
			}},
			CHECK: {Name: "check sessid*", Hand: func(m *ice.Message, arg ...string) { _sess_check(m, m.Option(ice.MSG_SESSID)) }},
		}, mdb.ImportantHashAction(mdb.EXPIRE, mdb.MONTH, mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,usernick,username,userrole,ua,ip"))},
	})
}

func SessCreate(m *ice.Message, username string) string {
	return m.Option(ice.MSG_SESSID, m.Cmdx(SESS, mdb.CREATE, username))
}
func SessCheck(m *ice.Message, sessid string) bool {
	m.Options(ice.MSG_USERNICK, "", ice.MSG_USERNAME, "", ice.MSG_USERROLE, VOID, "aaa.checker", logs.FileLine(-1))
	return sessid != "" && m.Cmdy(SESS, CHECK, sessid, logs.FileLineMeta(-1)).Option(ice.MSG_USERNAME) != ""
}
func SessAuth(m *ice.Message, value ice.Any, arg ...string) *ice.Message {
	switch val := value.(type) {
	case []string:
		value = kit.Dict(USERNICK, kit.Select("", val, 0), USERNAME, kit.Select("", val, 1), USERROLE, kit.Select("", val, 2))
	}
	return m.Auth(
		USERNICK, m.Option(ice.MSG_USERNICK, kit.Value(value, USERNICK)),
		USERNAME, m.Option(ice.MSG_USERNAME, kit.Value(value, USERNAME)),
		USERROLE, m.Option(ice.MSG_USERROLE, kit.Value(value, USERROLE)),
		arg, logs.FileLineMeta(kit.Select(logs.FileLine(-1), m.Option("aaa.checker"))),
	)
}
func SessLogout(m *ice.Message, arg ...string) {
	kit.If(m.Option(ice.MSG_SESSID) != "", func() { m.Cmd(SESS, mdb.REMOVE, mdb.HASH, m.Option(ice.MSG_SESSID)) })
}
