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

const (
	BACKGROUND = "background"
	AVATAR_URL = "avatar_url"
	AVATAR     = "avatar"
	GENDER     = "gender"
	MOBILE     = "mobile"
	SECRET     = "secret"
	THEME      = "theme"

	LANGUAGE  = "language"
	LOCATION  = "location"
	LONGITUDE = "longitude"
	LATITUDE  = "latitude"
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
)
const USER = "user"

func init() {
	Index.MergeCommands(ice.Commands{
		USER: {Help: "用户", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case USERNICK:
					m.Push(arg[0], m.Option(ice.MSG_USERNICK))
				case USERNAME:
					m.Push(arg[0], m.Option(ice.MSG_USERNAME))
				}
			}},
			mdb.CREATE: {Name: "create usernick username* userrole=void,tech userzone language", Hand: func(m *ice.Message, arg ...string) {
				_user_create(m, m.Option(USERNAME), m.OptionSimple(USERNICK, USERROLE, USERZONE, LANGUAGE, EMAIL, BACKGROUND, AVATAR)...)
			}},
		}, mdb.ImportantHashAction(mdb.SHORT, USERNAME, mdb.FIELD, "time,userrole,username,usernick,avatar,language,userzone", html.CHECKBOX, ice.TRUE))},
	})
}

func UserInfo(m *ice.Message, name ice.Any, key, meta string) (value string) {
	if m.Cmd(USER, kit.Select(m.Option(ice.MSG_USERNAME), name), func(val ice.Maps) { value = val[key] }).Length() == 0 || value == "" {
		return m.Option(meta)
	}
	return
}
func UserEmail(m *ice.Message, username ice.Any) (nick string) {
	return UserInfo(m, username, EMAIL, EMAIL)
}
func UserNick(m *ice.Message, username ice.Any) (nick string) {
	return UserInfo(m, username, USERNICK, ice.MSG_USERNICK)
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
func UserZone(m *ice.Message, username ice.Any) (zone string) {
	return UserInfo(m, username, USERZONE, ice.MSG_USERZONE)
}
func UserRoot(m *ice.Message, arg ...string) *ice.Message {
	language := kit.Select("", arg, 4)
	userzone := kit.Select("", arg, 3)
	userrole := kit.Select(ROOT, arg, 2)
	username := kit.Select(ice.Info.Username, arg, 1)
	usernick := kit.Select(UserNick(m, username), arg, 0)
	if len(arg) > 0 {
		ice.Info.Username = username
		m.Cmd(USER, mdb.CREATE, usernick, username, userrole, userzone, language)
	}
	return SessAuth(m, kit.Dict(USERNICK, usernick, USERNAME, username, USERROLE, userrole))
}
