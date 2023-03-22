package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _sess_create(m *ice.Message, username string, arg ...string) (h string) {
	if msg := m.Cmd(USER, username); msg.Length() > 0 {
		h = mdb.HashCreate(m, msg.AppendSimple(USERNAME, USERNICK, USERROLE), arg)
	} else {
		h = mdb.HashCreate(m, m.OptionSimple(USERNAME, USERNICK, USERROLE), arg)
	}
	gdb.Event(m, SESS_CREATE, SESS, h, USERNAME, username)
	return
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
	GRANT  = "grant"
	LOGIN  = "login"
	LOGOUT = "logout"
	SESSID = "sessid"
)
const (
	SESS_CREATE = "sess.create"
)
const SESS = "sess"

func init() {
	Index.MergeCommands(ice.Commands{
		SESS: {Name: "sess hash auto prunes", Help: "会话", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create username*", Hand: func(m *ice.Message, arg ...string) {
				_sess_create(m, m.Option(USERNAME), UA, m.Option(ice.MSG_USERUA), IP, m.Option(ice.MSG_USERIP))
			}},
			CHECK: {Name: "check sessid*", Hand: func(m *ice.Message, arg ...string) {
				_sess_check(m, m.Option(SESSID))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,username,usernick,userrole,ua,ip", mdb.EXPIRE, "720h", mdb.ImportantDataAction()))},
	})
}

func SessCreate(m *ice.Message, username string) string {
	return m.Option(ice.MSG_SESSID, m.Cmdx(SESS, mdb.CREATE, username))
}
func SessCheck(m *ice.Message, sessid string) bool {
	m.Options(ice.MSG_USERNAME, "", ice.MSG_USERNICK, "", ice.MSG_USERROLE, VOID, "aaa.checker", logs.FileLine(-1))
	return sessid != "" && m.Cmdy(SESS, CHECK, sessid, logs.FileLineMeta(-1)).Option(ice.MSG_USERNAME) != ""
}
func SessAuth(m *ice.Message, value ice.Any, arg ...string) *ice.Message {
	return m.Auth(
		USERNICK, m.Option(ice.MSG_USERNICK, kit.Value(value, USERNICK)),
		USERNAME, m.Option(ice.MSG_USERNAME, kit.Value(value, USERNAME)),
		USERROLE, m.Option(ice.MSG_USERROLE, kit.Value(value, USERROLE)),
		arg, logs.FileLineMeta(kit.Select(logs.FileLine(-1), m.Option("aaa.checker"))),
	)
}
func SessLogout(m *ice.Message, arg ...string) {
	if m.Option(ice.MSG_SESSID) == "" {
		return
	}
	m.Cmd(SESS, mdb.REMOVE, kit.Dict(mdb.HASH, m.Option(ice.MSG_SESSID)))
}
