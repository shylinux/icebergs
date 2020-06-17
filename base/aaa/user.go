package aaa

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/toolkits"
)

func _user_list(m *ice.Message) {
	m.Richs(USER, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		m.Push(key, value, []string{kit.MDB_TIME, USERNAME, USERNODE})
	})
}
func _user_login(m *ice.Message, name, word string) (ok bool) {
	m.Richs(USER, nil, name, func(key string, value map[string]interface{}) {
		if value[PASSWORD] == "" {
			ok, value[PASSWORD] = true, word
		} else if value[PASSWORD] == word {
			ok = true
		}
	})
	return ok
}
func _user_create(m *ice.Message, name, word string) {
	// 创建用户
	m.Rich(USER, nil, kit.Dict(
		USERNAME, name, PASSWORD, word,
		USERNICK, name, USERNODE, m.Conf(ice.CLI_RUNTIME, "boot.hostname"),
	))
	m.Log_CREATE(USERNAME, name)
	m.Event(ice.USER_CREATE, name)
}

func UserRole(m *ice.Message, username string) string {
	if username == cli.UserName {
		return ROOT
	}
	return VOID
}
func UserLogin(m *ice.Message, username, password string) bool {
	if _user_login(m, username, password) {
		m.Option(ice.MSG_USERNAME, username)
		m.Option(ice.MSG_USERROLE, UserRole(m, username))
		m.Option(ice.MSG_SESSID, SessCreate(m, m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_USERROLE)))

		m.Log_AUTH(
			USERROLE, m.Option(ice.MSG_USERROLE),
			USERNAME, m.Option(ice.MSG_USERNAME),
			SESSID, m.Option(ice.MSG_SESSID),
		)
		return true
	}
	return false
}
func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			USER: {Name: "user", Help: "用户", Value: kit.Data(kit.MDB_SHORT, USERNAME)},
		},
		Commands: map[string]*ice.Command{
			USER: {Name: "user first|login", Help: "用户", Action: map[string]*ice.Action{
				kit.MDB_CREATE: {Name: "create username [password]", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_user_create(m, arg[0], kit.Select("", arg, 1))
				}},
				"login": {Name: "login username password", Help: "login", Hand: func(m *ice.Message, arg ...string) {
					_user_login(m, arg[0], arg[1])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					_user_list(m)
					return
				}

				switch arg[0] {
				case "first":
					// 超级用户
					if m.Richs(ice.AAA_USER, nil, "%", nil) == nil {
					}

				case "login":
					// 用户认证
					user := m.Richs(USER, nil, arg[1], nil)
					if word := kit.Select("", arg, 2); user == nil {
						nick := arg[1]
						if len(nick) > 8 {
							nick = nick[:8]
						}
						_user_create(m, arg[1], word)

					} else if word != "" {
						if !_user_login(m, arg[1], word) {
							m.Info("login fail user: %s", arg[1])
							break
						}
					}

					if m.Options(ice.MSG_SESSID) && m.Cmdx(ice.AAA_SESS, "check", m.Option(ice.MSG_SESSID)) == arg[1] {
						// 复用会话
						m.Echo(m.Option(ice.MSG_SESSID))
						break
					}

					// 创建会话
					m.Echo(m.Cmdx(ice.AAA_SESS, "create", arg[1]))
				}
			}},
		},
	}, nil)
}
