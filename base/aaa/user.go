package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

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
	gdb.Event(m, USER_CREATE, USER, name)
}
func _user_login(m *ice.Message, name, word string) {
	if m.Warn(name == "", ice.ErrNotValid, name) {
		return
	}
	if !mdb.HashSelectDetail(m.Spawn(), name, nil) {
		_user_create(m.Spawn(), name, word)
	}

	_source := logs.FileLineMeta(logs.FileLine(-1))
	mdb.HashSelectDetail(m.Spawn(), name, func(value ice.Map) {
		if m.Warn(word != "" && word != kit.Format(kit.Value(value, PASSWORD)), ice.ErrNotRight) {
			return
		}
		m.Auth(
			USERROLE, m.Option(ice.MSG_USERROLE, value[USERROLE]),
			USERNAME, m.Option(ice.MSG_USERNAME, value[USERNAME]),
			USERNICK, m.Option(ice.MSG_USERNICK, value[USERNICK]),
			_source,
		)
	})
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

	USER_CREATE = "user.create"

	INVITE = "invite"
)
const USER = "user"

func init() {
	Index.MergeCommands(ice.Commands{
		USER: {Name: "user username auto create", Help: "用户", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create username password userrole=void,tech usernick", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				_user_create(m, m.Option(USERNAME), m.Option(PASSWORD), m.OptionSimple(USERROLE, USERNICK)...)
			}},
			LOGIN: {Name: "login username password", Help: "登录", Hand: func(m *ice.Message, arg ...string) {
				_user_login(m, m.Option(USERNAME), m.Option(PASSWORD))
			}},
		}, mdb.HashSearchAction(mdb.SHORT, USERNAME, mdb.FIELD, "time,userrole,username,usernick,userzone"))},
	})
}

func UserRoot(m *ice.Message, arg ...string) *ice.Message { // password username userrole
	userrole := m.Option(ice.MSG_USERROLE, ROOT)
	username := m.Option(ice.MSG_USERNAME, kit.Select(ice.Info.UserName, arg, 1))
	usernick := m.Option(ice.MSG_USERNICK, kit.Select(UserNick(m, username), arg, 2))
	if len(arg) > 0 {
		m.Cmd(USER, mdb.CREATE, username, kit.Select("", arg, 0), userrole, usernick)
		ice.Info.UserName = username
	}
	return m
}
func UserInfo(m *ice.Message, name ice.Any, key, meta string) (value string) {
	if m.Cmd(USER, name, func(val ice.Maps) {
		value = val[key]
	}).Length() == 0 && kit.Format(name) == m.Option(ice.MSG_USERNAME) {
		return m.Option(meta)
	}
	return
}
func UserRole(m *ice.Message, username ice.Any) (role string) {
	if username == "" {
		return VOID
	}
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
func UserLogin(m *ice.Message, username, password string) bool {
	m.Option(ice.MSG_USERROLE, VOID)
	m.Option(ice.MSG_USERNAME, "")
	m.Option(ice.MSG_USERNICK, "")
	return m.Cmdy(USER, LOGIN, username, password).Option(ice.MSG_USERNAME) != ""
}
