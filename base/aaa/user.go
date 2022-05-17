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
		if ok = !m.Warn(word != "" && word != value[PASSWORD], ice.ErrNotRight); ok {
			m.Log_AUTH(
				USERROLE, m.Option(ice.MSG_USERROLE, value[USERROLE]),
				USERNAME, m.Option(ice.MSG_USERNAME, value[USERNAME]),
				USERNICK, m.Option(ice.MSG_USERNICK, value[USERNICK]),
			)
		}
	})
	return ok
}
func _user_create(m *ice.Message, role, name, word string) {
	if name == "" {
		return
	}
	if word == "" {
		if m.Richs(USER, nil, name, func(key string, value map[string]interface{}) {
			word = kit.Format(value[PASSWORD])
		}) == nil {
			word = kit.Hashs()
		}
	}
	m.Rich(USER, nil, kit.Dict(USERROLE, role, USERNAME, name, PASSWORD, word))
	m.Event(USER_CREATE, USER, name)
}
func _user_search(m *ice.Message, name, text string) {
	m.Richs(USER, nil, mdb.FOREACH, func(key string, value map[string]interface{}) {
		if value = kit.GetMeta(value); name == "" || name == value[USERNAME] {
			m.PushSearch(kit.SimpleKV("", value[USERROLE], value[USERNAME], value[USERNICK]), value)
		}
	})
}

func UserRoot(m *ice.Message, arg ...string) *ice.Message { // password username userrole
	username := m.Option(ice.MSG_USERNAME, kit.Select(ice.Info.UserName, arg, 1))
	userrole := m.Option(ice.MSG_USERROLE, kit.Select(ROOT, arg, 2))
	if len(arg) > 0 {
		_user_create(m, userrole, username, kit.Select("", arg, 0))
		ice.Info.UserName = username
	}
	return m
}
func UserRole(m *ice.Message, username interface{}) (role string) {
	if role = VOID; username == ice.Info.UserName {
		return ROOT
	}
	if m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		role = kit.Format(kit.GetMeta(value)[USERROLE])
	}) == nil && kit.Format(username) == m.Option(ice.MSG_USERNAME) {
		return m.Option(ice.MSG_USERROLE)
	}
	return
}
func UserNick(m *ice.Message, username interface{}) (nick string) {
	if m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		nick = kit.Format(kit.GetMeta(value)[USERNICK])
	}) == nil && kit.Format(username) == m.Option(ice.MSG_USERNAME) {
		return m.Option(ice.MSG_USERNICK)
	}
	return
}
func UserZone(m *ice.Message, username interface{}) (zone string) {
	m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		zone = kit.Format(kit.GetMeta(value)[USERZONE])
	})
	return
}
func UserLogin(m *ice.Message, username, password string) bool {
	return _user_login(m, username, password)
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
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == USER {
					_user_search(m, arg[1], kit.Select("", arg, 2))
				}
			}},
			mdb.CREATE: {Name: "create userrole=void,tech username password", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if !_user_exists(m, m.Option(USERNAME)) {
					_user_create(m, m.Option(USERROLE), m.Option(USERNAME), m.Option(PASSWORD))
				}
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	}})
}
