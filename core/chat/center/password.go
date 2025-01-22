package center

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

type password struct {
	change string `name:"change" help:"修改登录"`
	login  string `name:"login username* password*" help:"登录" role:"void"`
	list   string `name:"list refresh" help:"密码登录" role:"void"`
}

func (s password) Change(m *ice.Message, arg ...string) {
	m.Cmd(aaa.USER, mdb.MODIFY, aaa.USERNAME, m.Option(ice.MSG_USERNAME), aaa.PASSWORD, arg[1])
}
func (s password) Login(m *ice.Message, arg ...string) {
	if m.WarnNotValid(m.Option(aaa.PASSWORD) != m.Cmd(aaa.USER, m.Option(aaa.USERNAME)).Append(aaa.PASSWORD), aaa.PASSWORD) {
		return
	}
	web.RenderCookie(m.Message, aaa.SessCreate(m.Message, m.Option(aaa.USERNAME)))
}
func (s password) List(m *ice.Message, arg ...string) {
	if m.Option(ice.MSG_USERNAME) == "" {
		m.DisplayForm("username*", "password*", s.Login)
	} else {
		m.DisplayForm("password*", s.Change)
	}
}

func init() { ice.Cmd("web.chat.password", password{}) }
