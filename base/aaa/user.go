package aaa

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _user_create(m *ice.Message, name string, arg ...string) {
	mdb.HashCreate(m, USERNAME, name, arg)
	gdb.Event(m, USER_CREATE, USER, name)
}

const (
	BACKGROUND = "background"
	AVATAR     = "avatar"
	GENDER     = "gender"
	MOBILE     = "mobile"

	CITY     = "city"
	COUNTRY  = "country"
	PROVINCE = "province"
	LANGUAGE = "language"
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
		USER: {Name: "user username auto create", Help: "用户", Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case USERNICK:
					m.Push(arg[0], m.Option(ice.MSG_USERNICK))
				case USERNAME:
					m.Push(arg[0], m.Option(ice.MSG_USERNAME))
				}
			}},
			mdb.CREATE: {Name: "create usernick username* userrole=void,tech userzone background", Hand: func(m *ice.Message, arg ...string) {
				_user_create(m, m.Option(USERNAME), m.OptionSimple(USERNICK, USERROLE, USERZONE)...)
			}},
		}, mdb.HashAction(mdb.SHORT, USERNAME, mdb.FIELD, "time,usernick,username,userrole,userzone"), mdb.ImportantHashAction())},
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
func UserRole(m *ice.Message, username ice.Any) (role string) {
	if username == "" {
		return VOID
	}
	if role = VOID; username == ice.Info.Username {
		return ROOT
	}
	return UserInfo(m, username, USERROLE, ice.MSG_USERROLE)
}
func UserZone(m *ice.Message, username ice.Any) (zone string) {
	return UserInfo(m, username, USERZONE, ice.MSG_USERZONE)
}
func UserRoot(m *ice.Message, arg ...string) *ice.Message {
	userzone := kit.Select("", arg, 3)
	userrole := kit.Select(ROOT, arg, 2)
	username := kit.Select(ice.Info.Username, arg, 1)
	usernick := kit.Select(UserNick(m, username), arg, 0)
	background := kit.Select("usr/icons/background.jpg", UserInfo(m, username, BACKGROUND, ""))
	if len(arg) > 0 {
		ice.Info.Username = username
		m.Cmd(USER, mdb.CREATE, usernick, username, userrole, userzone, background)
	}
	return SessAuth(m, kit.Dict(USERNICK, usernick, USERNAME, username, USERROLE, userrole))
}
