package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _user_create(m *ice.Message, name, word string, arg ...string) {
	mdb.HashCreate(m, USERNAME, name, PASSWORD, kit.GetValid(
		func() string { return word },
		func() string { return m.CmdAppend(USER, name, PASSWORD) },
		func() string { return kit.Hashs() },
	), arg)
	gdb.Event(m, USER_CREATE, USER, name)
}
func _user_login(m *ice.Message, name, word string) {
	if val := mdb.HashSelectDetails(m.Spawn(), name, func(value ice.Map) bool {
		return !m.Warn(word != "" && word != kit.Format(value[PASSWORD]), ice.ErrNotValid)
	}); len(val) > 0 {
		SessAuth(m, val)
	}
}

const (
	BACKGROUND = "background"

	AVATAR = "avatar"
	GENDER = "gender"
	MOBILE = "mobile"

	CITY     = "city"
	COUNTRY  = "country"
	PROVINCE = "province"
	LANGUAGE = "language"
)
const (
	USERNAME = "username"
	PASSWORD = "password"
	USERNICK = "usernick"
	USERZONE = "userzone"
	USERROLE = "userrole"

	USER_CREATE = "user.create"
)
const USER = "user"

func init() {
	Index.MergeCommands(ice.Commands{
		USER: {Name: "user username auto create", Help: "用户", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case USERNAME:
					m.Push(arg[0], m.Option(ice.MSG_USERNAME))
				case USERNICK:
					m.Push(arg[0], m.Option(ice.MSG_USERNICK))
				}
			}},
			mdb.CREATE: {Name: "create username* password usernick userzone userrole=void,tech", Hand: func(m *ice.Message, arg ...string) {
				_user_create(m, m.Option(USERNAME), m.Option(PASSWORD), m.OptionSimple(USERNICK, USERZONE, USERROLE)...)
			}},
			LOGIN: {Name: "login username* password", Hand: func(m *ice.Message, arg ...string) {
				_user_login(m, m.Option(USERNAME), m.Option(PASSWORD))
			}},
		}, mdb.HashSearchAction(mdb.SHORT, USERNAME, mdb.FIELD, "time,username,usernick,userzone,userrole"))},
	})
}

func UserInfo(m *ice.Message, name ice.Any, key, meta string) (value string) {
	if m.Cmd(USER, name, func(val ice.Maps) { value = val[key] }).Length() == 0 && kit.Format(name) == m.Option(ice.MSG_USERNAME) {
		return m.Option(meta)
	}
	return
}
func UserNick(m *ice.Message, username ice.Any) (nick string) {
	return UserInfo(m, username, USERNICK, ice.MSG_USERNICK)
}
func UserZone(m *ice.Message, username ice.Any) (zone string) {
	return UserInfo(m, username, USERZONE, ice.MSG_USERZONE)
}
func UserRole(m *ice.Message, username ice.Any) (role string) {
	if username == "" {
		return VOID
	}
	if role = VOID; username == ice.Info.Username {
		return ROOT
	}
	return UserInfo(m, username, USERROLE, ice.MSG_USERROLE)
}
func UserLogin(m *ice.Message, username, password string) bool {
	m.Options(ice.MSG_USERNAME, "", ice.MSG_USERNICK, "", ice.MSG_USERROLE, VOID)
	return m.Cmdy(USER, LOGIN, username, password).Option(ice.MSG_USERNAME) != ""
}
func UserRoot(m *ice.Message, arg ...string) *ice.Message {
	username := kit.Select(ice.Info.Username, arg, 0)
	usernick := kit.Select(UserNick(m, username), arg, 1)
	userrole := kit.Select(ROOT, arg, 2)
	userzone := kit.Select("", arg, 3)
	if len(arg) > 0 {
		m.Cmd(USER, mdb.CREATE, username, "", usernick, userzone, userrole)
		ice.Info.Username = username
	}
	return SessAuth(m, kit.Dict(USERNAME, username, USERNICK, usernick, USERROLE, userrole))
}
