package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _sess_check(m *ice.Message, sessid string) {
	m.Option(ice.MSG_USERROLE, VOID)
	m.Option(ice.MSG_USERNAME, "")
	m.Option(ice.MSG_USERNICK, "")
	if sessid == "" {
		return
	}

	_source := logs.FileLineMeta(logs.FileLine(-1, 3))
	mdb.HashSelectDetail(m, sessid, func(value ice.Map) {
		if m.Warn(kit.Time(kit.Format(value[mdb.TIME])) < kit.Time(m.Time()), ice.ErrNotValid, sessid) {
			return // 会话超时
		}
		m.Log_AUTH(
			USERROLE, m.Option(ice.MSG_USERROLE, value[USERROLE]),
			USERNAME, m.Option(ice.MSG_USERNAME, value[USERNAME]),
			USERNICK, m.Option(ice.MSG_USERNICK, value[USERNICK]),
			_source,
		)
	})
}
func _sess_create(m *ice.Message, username string) string {
	if username == "" {
		return ""
	}
	msg := m.Cmd(USER, username)
	if msg.Length() > 0 {
		h := mdb.HashCreate(m, msg.AppendSimple(USERROLE, USERNAME, USERNICK), IP, m.Option(ice.MSG_USERIP), UA, m.Option(ice.MSG_USERUA)).Result()
		m.Event(SESS_CREATE, SESS, h, USERNAME, username)
		return h
	} else {
		h := mdb.HashCreate(m, m.OptionSimple(USERROLE, USERNAME, USERNICK), IP, m.Option(ice.MSG_USERIP), UA, m.Option(ice.MSG_USERUA)).Result()
		m.Event(SESS_CREATE, SESS, h, USERNAME, username)
		return h
	}
}

const (
	IP = "ip"
	UA = "ua"
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
		SESS: {Name: "sess hash auto prunes", Help: "会话", Actions: ice.MergeAction(ice.Actions{
			mdb.CREATE: {Name: "create username", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				_sess_create(m, m.Option(USERNAME))
			}},
			CHECK: {Name: "check sessid", Help: "检查", Hand: func(m *ice.Message, arg ...string) {
				_sess_check(m, m.Option(SESSID))
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.UNIQ, mdb.FIELD, "time,hash,userrole,username,usernick,ip,ua", mdb.EXPIRE, "720h"))},
	})
}

func SessCreate(m *ice.Message, username string) string {
	return m.Option(ice.MSG_SESSID, m.Cmdx(SESS, mdb.CREATE, username))
}
func SessCheck(m *ice.Message, sessid string) bool {
	return m.Cmdy(SESS, CHECK, sessid).Option(ice.MSG_USERNAME) != ""
}
