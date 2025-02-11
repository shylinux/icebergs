package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _sess_create(m *ice.Message, username string, arg ...string) {
	if msg := m.Cmd(USER, username); msg.Length() > 0 {
		mdb.HashCreate(m, msg.AppendSimple(USERROLE, USERNAME, USERNICK, AVATAR), arg)
	} else {
		mdb.HashCreate(m, m.OptionSimple(USERROLE, USERNAME, USERNICK, AVATAR), arg)
	}
}
func _sess_check(m *ice.Message, sessid string) {
	if val := mdb.HashSelectDetails(m, sessid, func(value ice.Map) bool {
		return kit.Format(value[mdb.TIME]) > m.Time()
	}); len(val) > 0 {
		SessAuth(m, val)
	}
}

const (
	IP = "ip"
	UA = "ua"
)
const (
	CHECK  = "check"
	LOGIN  = "login"
	LOGOUT = "logout"
)
const SESS = "sess"

func init() {
	Index.MergeCommands(ice.Commands{
		SESS: {Name: "sess hash auto", Help: "会话", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create username*", Hand: func(m *ice.Message, arg ...string) {
				_sess_create(m, m.Option(USERNAME), UA, m.Option(ice.MSG_USERUA), IP, m.Option(ice.MSG_USERIP))
			}},
			CHECK: {Name: "check sessid*", Hand: func(m *ice.Message, arg ...string) { _sess_check(m, m.Option(ice.MSG_SESSID)) }},
		}, mdb.ImportantHashAction("checkbox", ice.TRUE, mdb.EXPIRE, mdb.MONTH, mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,userrole,username,usernick,avatar,ip,ua"))},
	})
}

func SessCreate(m *ice.Message, username string) string {
	return m.Option(ice.MSG_SESSID, m.Cmdx(SESS, mdb.CREATE, username))
}
func SessCheck(m *ice.Message, sessid string) bool {
	m.Options(ice.MSG_USERNICK, "", ice.MSG_USERNAME, "", ice.MSG_USERROLE, VOID, ice.MSG_CHECKER, logs.FileLine(-1))
	return sessid != "" && m.Cmdy(SESS, CHECK, sessid, logs.FileLineMeta(-1)).Option(ice.MSG_USERNAME) != ""
}
func SessValid(m *ice.Message) string {
	if m.Option(ice.MSG_SESSID) == "" || m.Spawn().AdminCmd(SESS, m.Option(ice.MSG_SESSID)).Length() == 0 {
		return m.Option(ice.MSG_SESSID, m.Spawn().AdminCmd(SESS, mdb.CREATE, m.Option(ice.MSG_USERNAME)).Result())
	}
	return m.Option(ice.MSG_SESSID)
}
func SessAuth(m *ice.Message, value ice.Any, arg ...string) *ice.Message {
	return m.Auth(
		USERROLE, m.Option(ice.MSG_USERROLE, kit.Format(kit.Value(value, USERROLE))),
		USERNAME, m.Option(ice.MSG_USERNAME, kit.Format(kit.Value(value, USERNAME))),
		USERNICK, m.Option(ice.MSG_USERNICK, kit.Format(kit.Value(value, USERNICK))),
		LANGUAGE, m.OptionDefault(ice.MSG_LANGUAGE, kit.Format(kit.Value(value, LANGUAGE))),
		AVATAR, m.Option(ice.MSG_AVATAR, kit.Format(kit.Value(value, AVATAR))),
		arg, logs.FileLineMeta(kit.Select(logs.FileLine(-1), m.Option("aaa.checker"))),
	)
}
func SessLogout(m *ice.Message, arg ...string) {
	kit.If(m.Option(ice.MSG_SESSID), func(sessid string) { m.Cmd(SESS, mdb.REMOVE, mdb.HASH, sessid) })
}
