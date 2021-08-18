package aaa

import (
	"strings"

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
				USERROLE, m.Option(ice.MSG_USERROLE, UserRole(m, value[USERNAME])),
				USERNICK, m.Option(ice.MSG_USERNICK, value[USERNICK]),
				USERNAME, m.Option(ice.MSG_USERNAME, value[USERNAME]),
			)
		})
	})
}
func _sess_create(m *ice.Message, username string) string {
	if m.Richs(USER, nil, username, nil) == nil {
		_user_create(m, username, kit.Hashs())
	}

	h := m.Cmdx(mdb.INSERT, SESS, "", mdb.HASH,
		USERNAME, username, USERROLE, UserRole(m, username),
		kit.MDB_TIME, m.Time(m.Conf(SESS, kit.Keym(kit.MDB_EXPIRE))),
		IP, m.Option(ice.MSG_USERIP), UA, m.Option(ice.MSG_USERUA),
	)
	m.Log_CREATE(SESS_CREATE, username, kit.MDB_HASH, h)
	m.Event(SESS_CREATE, username)
	return h
}

func SessCheck(m *ice.Message, sessid string) {
	_sess_check(m, sessid)
}
func SessCreate(m *ice.Message, username string) string {
	if username == "" {
		return ""
	}
	return m.Option(ice.MSG_SESSID, _sess_create(m, username))
}
func SessIsCli(m *ice.Message) bool {
	if m.Option(ice.MSG_USERUA) == "" || !strings.HasPrefix(m.Option(ice.MSG_USERUA), "Mozilla/5.0") {
		return true
	}
	return false
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
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SESS: {Name: SESS, Help: "会话", Value: kit.Data(kit.MDB_SHORT, "uniq", kit.MDB_EXPIRE, "720h")},
		},
		Commands: map[string]*ice.Command{
			SESS: {Name: "sess hash auto prunes", Help: "会话", Action: map[string]*ice.Action{
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,username,userrole,ip,ua")
					m.Cmdy(mdb.DELETE, SESS, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes before@date", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,username,userrole,ip,ua")
					m.Cmd(mdb.PRUNES, SESS, "", mdb.HASH, func(key string, value map[string]interface{}) bool {
						if value = kit.GetMeta(value); kit.Time(kit.Format(value[kit.MDB_TIME])) < kit.Time(m.Option("before")) {
							m.Push(key, value, kit.Split(m.Option(mdb.FIELDS)))
							return true
						}
						return false
					})
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,hash,username,userrole,ip")
				m.Cmdy(mdb.SELECT, SESS, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
