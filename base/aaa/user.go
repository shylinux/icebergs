package aaa

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"strings"
)

func _user_list(m *ice.Message) {
	m.Richs(USER, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		m.Push(key, value, []string{kit.MDB_TIME, USERNICK, USERNAME})
	})
	m.Sort(USERNAME)
}
func _user_login(m *ice.Message, name, word string) (ok bool) {
	if m.Richs(USER, nil, name, nil) == nil {
		_user_create(m, name, "")
	}

	m.Richs(USER, nil, name, func(key string, value map[string]interface{}) {
		if value[PASSWORD] == "" {
			ok, value[PASSWORD] = true, word
		} else if value[PASSWORD] == word {
			m.Log_AUTH(USERNAME, name, PASSWORD, strings.Repeat("*", len(word)))
			ok = true
		}
	})
	return ok
}
func _user_modify(m *ice.Message, name string, arg ...string) {
	if m.Richs(USER, nil, name, nil) == nil {
		_user_create(m, name, "")
	}

	m.Richs(USER, nil, name, func(key string, value map[string]interface{}) {
		m.Log_MODIFY(USERNAME, name, arg)
		for i := 0; i < len(arg)-1; i += 2 {
			kit.Value(value, arg[i], arg[i+1])
		}
	})
}
func _user_create(m *ice.Message, name, word string) {
	m.Rich(USER, nil, kit.Dict(
		USERNAME, name, PASSWORD, word,
		USERNICK, name, USERNODE, cli.NodeName,
	))
	m.Log_CREATE(USERNAME, name)
	m.Event(gdb.USER_CREATE, name)
}
func _user_search(m *ice.Message, kind, name, text string, arg ...string) {
	m.Richs(USER, nil, kit.MDB_FOREACH, func(key string, val map[string]interface{}) {
		if name != "" && name != val[USERNAME] {
			return
		}
		m.Push("pod", m.Option("pod"))
		m.Push("ctx", "aaa")
		m.Push("cmd", USER)
		m.Push(key, val, []string{kit.MDB_TIME})
		m.Push(kit.MDB_SIZE, kit.Format(""))
		m.Push(kit.MDB_TYPE, kit.Format(UserRole(m, val[USERNAME])))
		m.Push(kit.MDB_NAME, kit.Format(val[USERNICK]))
		m.Push(kit.MDB_TEXT, kit.Format(val[USERNAME]))
	})
}

func UserRoot(m *ice.Message) {
	cli.PassWord = kit.Hashs("uniq")
	cli.PassWord = cli.UserName
	_user_create(m, cli.UserName, cli.PassWord)
}
func UserNick(m *ice.Message, username interface{}) (nick string) {
	m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		nick = kit.Format(value[USERNICK])
	})
	return
}
func UserRole(m *ice.Message, username interface{}) (role string) {
	if role = VOID; username == cli.UserName {
		return ROOT
	}
	m.Richs(ROLE, nil, TECH, func(key string, value map[string]interface{}) {
		if kit.Value(value, kit.Keys(USER, username)) == true {
			role = TECH
		}
	})
	return
}
func UserLogin(m *ice.Message, username, password string) bool {
	if _user_login(m, username, password) {
		m.Option(ice.MSG_USERNAME, username)
		m.Option(ice.MSG_USERROLE, UserRole(m, username))
		m.Info("%s: %s", m.Option(ice.MSG_USERROLE), m.Option(ice.MSG_USERNAME))
		return true
	}
	return false
}

const USER = "user"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			USER: {Name: "user", Help: "用户", Value: kit.Data(kit.MDB_SHORT, USERNAME)},
		},
		Commands: map[string]*ice.Command{
			USER: {Name: "user", Help: "用户", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create username [password]", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_user_create(m, arg[0], kit.Select("", arg, 1))
				}},
				mdb.MODIFY: {Name: "modify username [key value]...", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) == 2 {
						_user_modify(m, m.Option("username"), arg[0], arg[1])
						return
					}
					_user_modify(m, arg[0], arg[1:]...)
				}},
				mdb.SEARCH: {Name: "search type name text arg...", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_user_search(m, arg[0], arg[1], kit.Select("", arg, 2))
				}},
				"login": {Name: "login username password", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
					_user_login(m, arg[0], arg[1])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { _user_list(m) }},
		},
	}, nil)
}
