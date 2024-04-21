package windows

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"

	wapi "github.com/iamacarpet/go-win64api"
)

type service struct {
	list string `name:"list name auto filter"`
}

func (s service) List(m *ice.Message, arg ...string) {
	list, err := wapi.GetServices()
	ListPush(m, list, err, cli.PID, "statusText", mdb.NAME, "displayName")
	m.RenameAppend("statusText", mdb.STATUS, "displayName", mdb.TEXT)
	m.StatusTimeCountStats(mdb.STATUS)
	m.Sort("status,name")
}

func init() { ice.ChatCtxCmd(service{}) }
