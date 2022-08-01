package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

func _user_login(m *ice.Message, name, word string) {
	if m.Warn(name == "", ice.ErrNotValid, name) {
		return
	}
	if !mdb.HashSelectDetail(m, name, nil) {
		_user_create(m, VOID, name, word)
	}

	_source := logs.FileLineMeta(logs.FileLine(-1, 3))
	mdb.HashSelectDetail(m, name, func(value ice.Map) {
		if m.Warn(word != "" && word != kit.Format(kit.Value(value, kit.Keys(mdb.EXTRA, PASSWORD))), ice.ErrNotRight) {
			return
		}
		m.Log_AUTH(
			USERROLE, m.Option(ice.MSG_USERROLE, value[USERROLE]),
			USERNAME, m.Option(ice.MSG_USERNAME, value[USERNAME]),
			USERNICK, m.Option(ice.MSG_USERNICK, value[USERNICK]),
			_source,
		)
	})
}
func _user_create(m *ice.Message, name, word string, arg ...string) {
	if m.Warn(name == "", ice.ErrNotValid, name) {
		return
	}
	if word == "" {
		word = m.CmdAppend(USER, name, PASSWORD)
	}
	if word == "" {
		word = kit.Hashs()
	}
	mdb.HashCreate(m, USERNAME, name, PASSWORD, word, arg)
	m.Event(USER_CREATE, USER, name)
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
	Index.MergeCommands(ice.Commands{
		USER: {Name: "user username auto create", Help: "用户", Actions: ice.MergeAction(ice.Actions{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectSearch(m, arg)
			}},
			mdb.CREATE: {Name: "create username password userrole=void,tech", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				_user_create(m, m.Option(USERNAME), m.Option(PASSWORD), m.OptionSimple(USERROLE)...)
			}},
			LOGIN: {Name: "login username password", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				_user_login(m, m.Option(USERNAME), m.Option(PASSWORD))
			}},
		}, mdb.HashAction(mdb.SHORT, USERNAME, mdb.FIELD, "time,userrole,username,usernick,userzone"))},
	})
}

func UserRoot(m *ice.Message, arg ...string) *ice.Message { // password username userrole
	username := m.Option(ice.MSG_USERNAME, kit.Select(ice.Info.UserName, arg, 1))
	userrole := m.Option(ice.MSG_USERROLE, kit.Select(ROOT, arg, 2))
	if len(arg) > 0 {
		m.Cmd(USER, mdb.CREATE, username, kit.Select("", arg, 0), userrole)
		ice.Info.UserName = username
	}
	return m
}
func UserRole(m *ice.Message, username ice.Any) (role string) {
	if role = VOID; username == ice.Info.UserName {
		return ROOT
	}
	return UserInfo(m, username, USERROLE, ice.MSG_USERROLE)
}
func UserNick(m *ice.Message, username ice.Any) (nick string) {
	return UserInfo(m, username, USERNICK, ice.MSG_USERNICK)
}
func UserZone(m *ice.Message, username ice.Any) (zone string) {
	return UserInfo(m, username, USERZONE, ice.MSG_USERZONE)
}
func UserInfo(m *ice.Message, name ice.Any, key, meta string) (value string) {
	if m.Cmd(USER, name).Tables(func(val ice.Maps) {
		value = val[key]
	}).Length() == 0 && kit.Format(name) == m.Option(ice.MSG_USERNAME) {
		return m.Option(meta)
	}
	return
}
func UserLogin(m *ice.Message, username, password string) bool {
	return m.Cmdy(USER, LOGIN, username, password).Option(ice.MSG_USERNAME) != ""
}
