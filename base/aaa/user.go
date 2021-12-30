package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _user_exists(m *ice.Message, name string) bool {
	return m.Richs(USER, nil, name, nil) != nil
}
func _user_login(m *ice.Message, name, word string) (ok bool) {
	if !_user_exists(m, name) {
		_user_create(m, VOID, name, word)
	}

	m.Richs(USER, nil, name, func(key string, value map[string]interface{}) {
		ok = word == value[PASSWORD]
	})
	return ok
}
func _user_create(m *ice.Message, role, name, word string) {
	if word == "" {
		word = kit.Hashs()
	}
	m.Rich(USER, nil, kit.Dict(
		USERROLE, role, USERNAME, name, PASSWORD, word,
		USERNICK, name, USERZONE, m.Option(ice.MSG_USERZONE),
	))
	m.Event(USER_CREATE, USER, name)
}
func _user_search(m *ice.Message, name, text string) {
	m.Richs(USER, nil, mdb.FOREACH, func(key string, value map[string]interface{}) {
		if value = kit.GetMeta(value); name != "" && name != value[USERNAME] {
			return
		}
		m.PushSearch(kit.SimpleKV("type,name,text",
			kit.Format(value[USERROLE]), kit.Format(value[USERNAME]), kit.Format(value[USERNICK])), value)
	})
}

func UserRoot(m *ice.Message) {
	m.Option(ice.MSG_USERROLE, ROOT)
	m.Option(ice.MSG_USERNAME, ice.Info.UserName)

	if !_user_exists(m, ice.Info.UserName) {
		_user_create(m, ROOT, ice.Info.UserName, "")
	}
}
func UserRole(m *ice.Message, username interface{}) (role string) {
	if role = VOID; username == ice.Info.UserName {
		return ROOT
	}
	m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		role = kit.Format(kit.GetMeta(value)[USERROLE])
	})
	return
}
func UserNick(m *ice.Message, username interface{}) (nick string) {
	m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		nick = kit.Format(kit.GetMeta(value)[USERNICK])
	})
	return
}
func UserZone(m *ice.Message, username interface{}) (zone string) {
	m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		zone = kit.Format(kit.GetMeta(value)[USERZONE])
	})
	return
}
func UserLogin(m *ice.Message, username, password string) bool {
	if password == "" || _user_login(m, username, password) {
		m.Log_AUTH(
			USERROLE, m.Option(ice.MSG_USERROLE, UserRole(m, username)),
			USERNAME, m.Option(ice.MSG_USERNAME, username),
			USERNICK, m.Option(ice.MSG_USERNICK, UserNick(m, username)),
		)
		return true
	}
	return false
}

const (
	BACKGROUND = "background"

	AVATAR = "avatar"
	GENDER = "gender"
	MOBILE = "mobile"
	EMAIL  = "email"

	CITY     = "city"
	COUNTRY  = "country"
	LANGUAGE = "language"
	PROVINCE = "province"
)
const (
	USERROLE = "userrole"
	USERNAME = "username"
	PASSWORD = "password"
	USERNICK = "usernick"
	USERZONE = "userzone"
)
const (
	USER_CREATE = "user.create"
	USER_REMOVE = "user.remove"
)
const (
	INVITE = "invite"
)
const USER = "user"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		USER: {Name: USER, Help: "用户", Value: kit.Data(
			mdb.SHORT, USERNAME, mdb.FIELD, "time,userrole,username,usernick,userzone",
		)},
	}, Commands: map[string]*ice.Command{
		USER: {Name: "user username auto create", Help: "用户", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, USER, m.PrefixKey())
			}},
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == USER {
					_user_search(m, arg[1], kit.Select("", arg, 2))
				}
			}},
			mdb.CREATE: {Name: "create userrole=void,tech username password", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if _user_exists(m, m.Option(USERNAME)) {
					return
				}
				_user_create(m, m.Option(USERROLE), m.Option(USERNAME), m.Option(PASSWORD))
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	}})
}
