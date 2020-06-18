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
		USERNICK, name, USERNODE, cli.NodeName,
	))
	m.Log_CREATE(USERNAME, name)
	m.Event(ice.USER_CREATE, name)
}

func UserRoot(m *ice.Message) {
	cli.PassWord = kit.Hashs("uniq")
	cli.PassWord = cli.UserName
	_user_create(m, cli.UserName, cli.PassWord)
}
func UserRole(m *ice.Message, username interface{}) string {
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
				_user_list(m)
			}},
		},
	}, nil)
}
