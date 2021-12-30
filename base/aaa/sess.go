package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _sess_check(m *ice.Message, sessid string) {
	m.Richs(SESS, nil, sessid, func(value map[string]interface{}) {
		value = kit.GetMeta(value)
		m.Richs(USER, nil, value[USERNAME], func(value map[string]interface{}) {
			value = kit.GetMeta(value)

			m.Log_AUTH(
				USERROLE, m.Option(ice.MSG_USERROLE, value[USERROLE]),
				USERNAME, m.Option(ice.MSG_USERNAME, value[USERNAME]),
				USERNICK, m.Option(ice.MSG_USERNICK, value[USERNICK]),
			)
		})
	})
}
func _sess_create(m *ice.Message, username string) string {
	if username == "" {
		return ""
	}
	if m.Richs(USER, nil, username, nil) == nil {
		_user_create(m, kit.Select(TECH, VOID, m.Option(ice.MSG_USERROLE) == VOID), username, kit.Hashs())
	}

	h := m.Cmdx(mdb.INSERT, SESS, "", mdb.HASH,
		USERROLE, UserRole(m, username), USERNAME, username,
		IP, m.Option(ice.MSG_USERIP), UA, m.Option(ice.MSG_USERUA),
		mdb.TIME, m.Time(m.Conf(SESS, kit.Keym(mdb.EXPIRE))),
	)
	m.Event(SESS_CREATE, SESS, h, USERNAME, username)
	return h
}

func SessCreate(m *ice.Message, username string) string {
	return m.Option(ice.MSG_SESSID, _sess_create(m, username))
}
func SessCheck(m *ice.Message, sessid string) {
	_sess_check(m, sessid)
}

const (
	IP = "ip"
	UA = "ua"
)
const (
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
