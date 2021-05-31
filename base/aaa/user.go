package aaa

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _user_login(m *ice.Message, name, word string) (ok bool) {
	if m.Richs(USER, nil, name, nil) == nil {
		_user_create(m, name, word)
	}

	m.Richs(USER, nil, name, func(key string, value map[string]interface{}) {
		if value[PASSWORD] == word {
			ok = true
		}
	})
	return ok
}
func _user_create(m *ice.Message, name, word string) {
	h := m.Rich(USER, nil, kit.Dict(
		USERNAME, name, PASSWORD, word,
		USERNICK, name, USERZONE, m.Option(ice.MSG_USERZONE),
	))
	m.Log_CREATE(USER_CREATE, name, kit.MDB_HASH, h)
	m.Event(USER_CREATE, name)
}
func _user_remove(m *ice.Message, name string) {
	m.Cmdy(mdb.DELETE, USER, "", mdb.HASH, USERNAME, name)
	m.Log_REMOVE(USER_REMOVE, name, kit.MDB_HASH, kit.Hashs(name))
	m.Event(USER_REMOVE, name)
}
func _user_search(m *ice.Message, kind, name, text string) {
	m.Richs(USER, nil, kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		if value = kit.GetMeta(value); name != "" && name != value[USERNAME] {
			return
		}
		m.PushSearch(kit.SSH_CMD, USER, kit.MDB_TYPE, kit.Format(UserRole(m, value[USERNAME])),
			kit.MDB_NAME, kit.Format(value[USERNICK]), kit.MDB_TEXT, kit.Format(value[USERNAME]), value)
	})
}

func UserRoot(m *ice.Message) {
	m.Option(ice.MSG_USERNAME, ice.Info.UserName)
	m.Option(ice.MSG_USERROLE, ROOT)

	if m.Richs(USER, "", ice.Info.UserName, nil) == nil {
		_user_create(m, ice.Info.UserName, kit.Hashs())
	}
}
func UserZone(m *ice.Message, username interface{}) (zone string) {
	m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		value = kit.GetMeta(value)
		zone = kit.Format(value[USERZONE])
	})
	return
}
func UserNick(m *ice.Message, username interface{}) (nick string) {
	m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		value = kit.GetMeta(value)
		nick = kit.Format(value[USERNICK])
	})
	return
}
func UserRole(m *ice.Message, username interface{}) (role string) {
	if role = VOID; username == ice.Info.UserName {
		return ROOT
	}
	m.Richs(ROLE, nil, TECH, func(key string, value map[string]interface{}) {
		if kit.Value(kit.GetMeta(value), kit.Keys(USER, username)) == true {
			role = TECH
		}
	})
	return
}
func UserLogin(m *ice.Message, username, password string) bool {
	if password == "" || _user_login(m, username, password) {
		m.Log_AUTH(
			USERNICK, m.Option(ice.MSG_USERNICK, UserNick(m, username)),
			USERROLE, m.Option(ice.MSG_USERROLE, UserRole(m, username)),
			USERNAME, m.Option(ice.MSG_USERNAME, username),
		)
		return true
	}
	return false
}

const (
	INVITE = "invite"
)
const (
	AVATAR = "avatar"
	GENDER = "gender"
	MOBILE = "mobile"
	EMAIL  = "email"

	CITY     = "city"
	COUNTRY  = "country"
	PROVINCE = "province"
	LANGUAGE = "language"

	BACKGROUND = "background"
)
const (
	USERZONE = "userzone"
	USERNICK = "usernick"
	USERROLE = "userrole"
	USERNAME = "username"
	PASSWORD = "password"
)
const (
	USER_CREATE = "user.create"
	USER_REMOVE = "user.remove"
)
const USER = "user"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			USER: {Name: USER, Help: "用户", Value: kit.Data(kit.MDB_SHORT, USERNAME)},
		},
		Commands: map[string]*ice.Command{
			USER: {Name: "user username auto create", Help: "用户", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create userrole=void,tech username password", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_user_create(m, m.Option(USERNAME), m.Option(PASSWORD))
					_role_user(m, m.Option(USERROLE), m.Option(USERNAME))
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, USER, "", mdb.HASH, USERNAME, m.Option(USERNAME), arg)
				}},
				mdb.REMOVE: {Name: "remove username", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					_user_remove(m, m.Option(USERNAME))
				}},
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == USER {
						_user_search(m, arg[0], arg[1], kit.Select("", arg, 2))
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg) == 0, "time,username,userzone,usernick")
				m.Cmdy(mdb.SELECT, USER, "", mdb.HASH, USERNAME, arg)
				m.Table(func(index int, value map[string]string, head []string) {
					m.Push(USERROLE, UserRole(m, value[USERNAME]))
				})
				m.PushAction(mdb.REMOVE)
			}},
		},
	})
}
