package aaa

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _sess_auth(m *ice.Message, sessid string, username string, userrole string) {
	m.Richs(SESS, nil, sessid, func(value map[string]interface{}) {
		if value = kit.GetMeta(value); m.Option(ice.MSG_USERROLE) == ROOT {
			value[USERROLE] = userrole
		} else if m.Option(ice.MSG_USERROLE) == TECH && userrole != ROOT {
			value[USERROLE] = userrole
		} else {
			return
		}
		m.Log_AUTH(SESSID, sessid, USERNAME, username, USERROLE, userrole)
		value[USERNAME] = username
	})
}
func _sess_check(m *ice.Message, sessid string) {
	m.Richs(SESS, nil, sessid, func(value map[string]interface{}) {
		value = kit.GetMeta(value)
		m.Richs(USER, nil, value[USERNAME], func(value map[string]interface{}) {
			value = kit.GetMeta(value)

			if m.Option(ice.MSG_USERNICK, value[USERNICK]) == "" {
				if name := kit.Format(value[USERNAME]); len(name) > 10 {
					m.Option(ice.MSG_USERNICK, name[:10])
				} else {
					m.Option(ice.MSG_USERNICK, value[USERNAME])
				}
			}
			m.Log_AUTH(
				USERROLE, m.Option(ice.MSG_USERROLE, kit.Select(UserRole(m, value[USERNAME]))),
				USERNICK, m.Option(ice.MSG_USERNICK, value[USERNICK]),
				USERNAME, m.Option(ice.MSG_USERNAME, value[USERNAME]),
			)
		})
	})
}
func _sess_create(m *ice.Message, username string) string {
	if m.Richs(USER, nil, username, nil) == nil {
		_user_create(m, username, "")
	}

	m.Cmdy(mdb.INSERT, SESS, "", mdb.HASH,
		USERNAME, username, "from", m.Option(ice.MSG_SESSID),
		kit.MDB_TIME, m.Time(m.Conf(SESS, "meta.expire")),
		"agent", m.Option(ice.MSG_USERUA),
		"ip", m.Option(ice.MSG_USERIP),
	)
	return m.Result()
}

func SessCheck(m *ice.Message, sessid string) *ice.Message {
	_sess_check(m, sessid)
	return m
}
func SessCreate(m *ice.Message, username, userrole string) string {
	_sess_auth(m, _sess_create(m, username), username, userrole)
	return m.Result()
}

const SESS = "sess"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SESS: {Name: SESS, Help: "会话", Value: kit.Data(
				kit.MDB_SHORT, "uniq", "expire", "720h",
			)},
		},
		Commands: map[string]*ice.Command{
			SESS: {Name: "sess hash auto", Help: "会话", Action: map[string]*ice.Action{
				"auth": {Name: "auth sessid username [userrole]", Help: "授权", Hand: func(m *ice.Message, arg ...string) {
					_sess_auth(m, arg[0], arg[1], kit.Select("", arg, 2))
				}},
				"check": {Name: "check sessid", Help: "校验", Hand: func(m *ice.Message, arg ...string) {
					_sess_check(m, arg[0])
				}},

				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, SESS, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,username,userrole", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, SESS, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(mdb.REMOVE)
			}},
		},
	}, nil)
}
