package windows

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"

	wapi "github.com/iamacarpet/go-win64api"
)

type logged struct {
	list string `name:"list username auto"`
}

func (s logged) List(m *ice.Message, arg ...string) {
	list, err := wapi.ListLoggedInUsers()
	ListPush(m, list, err, "logonTime", aaa.USERNAME, web.DOMAIN, "isLocal", "isAdmin")
	m.RenameAppend("logonTime", mdb.TIME)
	m.RewriteAppend(func(value, key string, index int) string {
		kit.If(key == mdb.TIME, func() { value = ParseTime(m, value) })
		return value
	})
}

func init() { ice.ChatCtxCmd(logged{}) }
