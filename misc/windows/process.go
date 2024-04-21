package windows

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"

	wapi "github.com/iamacarpet/go-win64api"
)

type process struct {
	list string `name:"list name auto filter"`
}

func (s process) List(m *ice.Message, arg ...string) {
	list, err := wapi.ProcessList()
	ListPush(m, list, err, "parentpid", cli.PID, aaa.USERNAME, "exeName", "fullPath")
	m.RenameAppend("parentpid", cli.PPID, "exeName", mdb.NAME, "fullPath", nfs.PATH)
	m.Sort("username,path")
}

func init() { ice.ChatCtxCmd(process{}) }
