package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _sess_check(m *ice.Message, sessid string) bool {
	m.Option(ice.MSG_USERROLE, VOID)
	m.Option(ice.MSG_USERNAME, "")
	m.Option(ice.MSG_USERNICK, "")
	if sessid == "" {
		return false
	}

	m.Richs(SESS, nil, sessid, func(value map[string]interface{}) {
		if value = kit.GetMeta(value); m.Warn(kit.Time(kit.Format(value[mdb.TIME])) < kit.Time(m.Time()), ice.ErrExpire) {
			return // 会话超时
		}
		if m.Richs(USER, nil, value[USERNAME], func(value map[string]interface{}) {
			m.Log_AUTH(
				USERROLE, m.Option(ice.MSG_USERROLE, value[USERROLE]),
				USERNAME, m.Option(ice.MSG_USERNAME, value[USERNAME]),
				USERNICK, m.Option(ice.MSG_USERNICK, value[USERNICK]),
			)
		}) == nil {
			m.Log_AUTH(
				USERROLE, m.Option(ice.MSG_USERROLE, value[USERROLE]),
				USERNAME, m.Option(ice.MSG_USERNAME, value[USERNAME]),
				USERNICK, m.Option(ice.MSG_USERNICK, value[USERNICK]),
			)
		}
	})
	return m.Option(ice.MSG_USERNAME) != ""
}
func _sess_create(m *ice.Message, username string) string {
	if username == "" {
		return ""
	}

	h := m.Cmdx(mdb.INSERT, SESS, "", mdb.HASH, mdb.TIME, m.Time(m.Conf(SESS, kit.Keym(mdb.EXPIRE))),
		USERROLE, UserRole(m, username), USERNAME, username, USERNICK, UserNick(m, username),
		IP, m.Option(ice.MSG_USERIP), UA, m.Option(ice.MSG_USERUA),
	)
	m.Event(SESS_CREATE, SESS, h, USERNAME, username)
	return h
}

func SessCreate(m *ice.Message, username string) string {
	return m.Option(ice.MSG_SESSID, _sess_create(m, username))
}
func SessCheck(m *ice.Message, sessid string) bool {
	return _sess_check(m, sessid)
}

const (
	IP = "ip"
	UA = "ua"
)
const (
	GRANT  = "grant"
	LOGIN  = "login"
	LOGOUT = "logout"
)
const (
	SESS_CREATE = "sess.create"
)
const SESS = "sess"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SESS: {Name: SESS, Help: "会话", Value: kit.Data(
			mdb.SHORT, "uniq", mdb.FIELD, "time,hash,userrole,username,ip,ua", mdb.EXPIRE, "720h",
		)},
	}, Commands: map[string]*ice.Command{
		SESS: {Name: "sess hash auto prunes", Help: "会话", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create username", Help: "创建"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	}})
}
