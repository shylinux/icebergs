package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _user_create(m *ice.Message, name string, arg ...string) {
	mdb.HashCreate(m, USERNAME, name, arg)
	gdb.Event(m, USER_CREATE, USER, name)
}
func _user_remove(m *ice.Message, name string, arg ...string) {
	gdb.Event(m, USER_REMOVE, m.OptionSimple(USERNAME, USERNICK))
	mdb.HashRemove(m, m.OptionSimple(USERNAME))
}

const (
	BACKGROUND = "background"
	AVATAR_URL = "avatar_url"
	AVATAR     = "avatar"
	GENDER     = "gender"
	MOBILE     = "mobile"
	PHONE      = "phone"
	SECRET     = "secret"
	THEME      = "theme"

	LANGUAGE  = "language"
	LOCATION  = "location"
	LONGITUDE = "longitude"
	LATITUDE  = "latitude"
	COMPANY   = "company"
	PROVINCE  = "province"
	COUNTRY   = "country"
	CITY      = "city"
)
const (
	USERNICK = "usernick"
	USERNAME = "username"
	PASSWORD = "password"
	USERROLE = "userrole"
	USERZONE = "userzone"

	USER_CREATE = "user.create"
	USER_REMOVE = "user.remove"
)
const USER = "user"

func init() {
	Index.MergeCommands(ice.Commands{
		USER: {Help: "用户", Icon: "Contacts.png", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case USERNICK:
					m.Push(arg[0], m.Option(ice.MSG_USERNICK))
				case USERNAME:
					m.Push(arg[0], m.Option(ice.MSG_USERNAME))
				}
			}},
			mdb.CREATE: {Name: "create userrole=void,tech username* usernick language userzone email", Hand: func(m *ice.Message, arg ...string) {
				_user_create(m, m.Option(USERNAME), m.OptionSimple(USERROLE, USERNICK, LANGUAGE, AVATAR, BACKGROUND, USERZONE, EMAIL)...)
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) { _user_remove(m, m.Option(USERNAME)) }},
		}, mdb.ImportantHashAction(mdb.SHORT, USERNAME, mdb.FIELD, "time,userrole,username,usernick,language,avatar,background,userzone", html.CHECKBOX, ice.TRUE))},
	})
}

func UserInfo(m *ice.Message, name ice.Any, key, meta string) (value string) {
	if m.Cmd(USER, kit.Select(m.Option(ice.MSG_USERNAME), name), func(val ice.Maps) { value = val[key] }).Length() == 0 || value == "" {
		return m.Option(meta)
	}
	return
}
func UserRole(m *ice.Message, username ice.Any) (role string) {
	if username == "" {
		return VOID
	} else if role = VOID; username == ice.Info.Username {
		return ROOT
	} else {
		return UserInfo(m, username, USERROLE, ice.MSG_USERROLE)
	}
}
func UserNick(m *ice.Message, username ice.Any) (nick string) {
	return UserInfo(m, username, USERNICK, ice.MSG_USERNICK)
}
func UserLang(m *ice.Message, username ice.Any) (nick string) {
	return UserInfo(m, username, LANGUAGE, ice.MSG_LANGUAGE)
}
func UserZone(m *ice.Message, username ice.Any) (zone string) {
	return UserInfo(m, username, USERZONE, ice.MSG_USERZONE)
}
func UserEmail(m *ice.Message, username ice.Any) (nick string) {
	return UserInfo(m, username, EMAIL, EMAIL)
}
func UserRoot(m *ice.Message, arg ...string) *ice.Message {
	userrole := kit.Select(TECH, arg, 0)
	username := kit.Select(ice.Info.Username, arg, 1)
	usernick := kit.Select(UserNick(m, username), arg, 2)
	language := kit.Select(UserLang(m, username), arg, 3)
	userzone := kit.Select(ice.OPS, arg, 4)
	email := kit.Select(UserEmail(m, username), arg, 5)
	if len(arg) > 0 {
		ice.Info.Username = username
		m.Cmd(USER, mdb.CREATE, userrole, username, usernick, language, userzone, email)
	}
	return SessAuth(m, kit.Dict(USERROLE, userrole, USERNAME, username, USERNICK, usernick))
}
