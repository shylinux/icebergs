package node

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/web"
)

type api struct {
	adminLogin  string `http:"/api/admin/login"`
	adminLogout string `http:"/api/admin/logout"`
	spaceList   string `http:"/api/space/list"`
	spaceAdd    string `http:"/api/space/add"`
	userList    string `http:"/api/user/list"`
	userAdd     string `http:"/api/user/add"`
}

func (s api) AdminLogin(m *ice.Message, arg ...string) {
}
func (s api) AdminLogout(m *ice.Message, arg ...string) {
}
func (s api) SpaceList(m *ice.Message, arg ...string) {
	m.Cmdy(web.DREAM)
}
func (s api) SpaceAdd(m *ice.Message, arg ...string) {
	m.Cmdy(web.DREAM)
}
func (s api) UserAdd(m *ice.Message, arg ...string) {
	m.Cmdy(aaa.USER)
}
func (s api) UserList(m *ice.Message, arg ...string) {
	m.Cmdy(aaa.USER)
}
func (s api) List() {

}

func init() { ice.CodeCtxCmd(api{}) }
