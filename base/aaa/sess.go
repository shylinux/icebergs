package aaa

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

func _sess_list(m *ice.Message) {
	m.Richs(SESS, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		m.Push(key, value, []string{kit.MDB_TIME, kit.MDB_KEY, USERNAME, USERROLE})
	})
}
func _sess_auth(m *ice.Message, sessid string, username string, userrole string) {
	m.Richs(SESS, nil, sessid, func(value map[string]interface{}) {
		if m.Option(ice.MSG_USERROLE) == ROOT {
			value[USERROLE] = userrole
		} else if m.Option(ice.MSG_USERROLE) == TECH && userrole != ROOT {
			value[USERROLE] = userrole
		} else {
			return
		}
		value["username"] = username
		m.Log_AUTH(SESSID, sessid, USERNAME, username, USERROLE, userrole)
	})
}
func _sess_check(m *ice.Message, sessid string) {
	m.Richs(SESS, nil, sessid, func(value map[string]interface{}) {
		m.Log_AUTH(
			USERROLE, m.Option(ice.MSG_USERROLE, value[USERROLE]),
			USERNAME, m.Option(ice.MSG_USERNAME, value[USERNAME]),
		)
	})
}
func _sess_create(m *ice.Message, username string) string {
	h := m.Rich(SESS, nil, kit.Dict(
		kit.MDB_TIME, m.Time(m.Conf(SESS, "meta.expire")),
		USERNAME, username, "from", m.Option(ice.MSG_SESSID),
	))
	m.Log_CREATE(SESSID, h, USERNAME, username)
	m.Echo(h)
	return h
}

func SessCheck(m *ice.Message, sessid string) *ice.Message {
	_sess_check(m, sessid)
	return m
}
func SessCreate(m *ice.Message, username, userrole string) string {
	_sess_auth(m, _sess_create(m, username), username, userrole)
	return m.Result()
}
func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SESS: {Name: "sess", Help: "会话", Value: kit.Data(
				kit.MDB_SHORT, "uniq", "expire", "720h",
			)},
		},
		Commands: map[string]*ice.Command{
			SESS: {Name: "sess check|login", Help: "会话", Action: map[string]*ice.Action{
				kit.MDB_CREATE: {Name: "create [username]", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_sess_create(m, arg[0])
				}},
				"check": {Name: "check sessid", Help: "校验", Hand: func(m *ice.Message, arg ...string) {
					_sess_check(m, arg[0])
				}},
				"auth": {Name: "auth sessid username [userrole]", Help: "授权", Hand: func(m *ice.Message, arg ...string) {
					_sess_auth(m, arg[0], arg[1], kit.Select("", arg, 2))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_sess_list(m)
			}},
		},
	}, nil)
}
