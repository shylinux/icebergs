package aaa

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"strings"
)

func _user_login(m *ice.Message, name, word string) (ok bool) {
	if m.Richs(USER, nil, name, nil) == nil {
		_user_create(m, name, "")
	}

	m.Richs(USER, nil, name, func(key string, value map[string]interface{}) {
		if kit.Format(value[PASSWORD]) == "" {
			ok, value[PASSWORD] = true, word
		} else if value[PASSWORD] == word {
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
	m.Log_CREATE(USERNAME, name, "hash", h)
	m.Event(USER_CREATE, name)
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
	ice.Info.PassWord = kit.Hashs("uniq")
	ice.Info.PassWord = ice.Info.UserName
	_user_create(m, ice.Info.UserName, ice.Info.PassWord)
}
func UserNick(m *ice.Message, username interface{}) (nick string) {
	m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		value = kit.GetMeta(value)
		nick = kit.Format(value[USERNICK])
	})
	return
}
func UserZone(m *ice.Message, username interface{}) (zone string) {
	m.Richs(USER, nil, kit.Format(username), func(key string, value map[string]interface{}) {
		value = kit.GetMeta(value)
		zone = kit.Format(value[USERZONE])
	})
	return
}
func UserRole(m *ice.Message, username interface{}) (role string) {
	if role = VOID; username == ice.Info.UserName {
		return ROOT
	}
	m.Richs(ROLE, nil, TECH, func(key string, value map[string]interface{}) {
		value = kit.GetMeta(value)
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
		m.Log_AUTH(USERROLE, m.Option(ice.MSG_USERROLE), USERNAME, m.Option(ice.MSG_USERNAME), PASSWORD, strings.Repeat("*", len(password)))
		return true
	}
	return false
}

const (
	USER_CREATE = "user.create"
)

const USER = "user"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			USER: {Name: USER, Help: "用户", Value: kit.Data(kit.MDB_SHORT, USERNAME)},
		},
		Commands: map[string]*ice.Command{
			USER: {Name: "user username auto", Help: "用户", Action: map[string]*ice.Action{
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, USER, "", mdb.HASH, USERNAME, m.Option(USERNAME), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.REMOVE, USER, "", mdb.HASH, USERNAME, m.Option(USERNAME))
				}},
				mdb.SEARCH: {Name: "search type name text arg...", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_user_search(m, arg[0], arg[1], kit.Select("", arg, 2))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,userzone,usernick,username", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, USER, "", mdb.HASH, USERNAME, arg)
				m.PushAction(mdb.REMOVE)
			}},
		},
	}, nil)
}
