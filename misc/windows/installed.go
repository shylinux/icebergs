package windows

import (
	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"

	wapi "github.com/iamacarpet/go-win64api"
)

type installed struct {
	list string `name:"list name auto filter"`
}

func (s installed) List(m *ice.Message, arg ...string) {
	list, err := wapi.InstalledSoftwareList()
	ListPush(m, list, err,
		"installDate",
		"arch", "estimatedSize",
		"displayName", "displayVersion",
		"publisher", "HelpLink",
		"InstallLocation",
		"UninstallString",
	)
	m.RenameAppend(
		"installDate", mdb.TIME,
		"estimatedSize", nfs.SIZE,
		"displayName", mdb.NAME,
		"displayVersion", nfs.VERSION,
	)
	m.RewriteAppend(func(value, key string, index int) string {
		kit.If(key == nfs.SIZE, func() { value = kit.FmtSize(kit.Int(value) * 1024) })
		kit.If(key == mdb.TIME, func() { value = ParseTime(m, value) })
		return value
	})
	m.SortIntR(nfs.SIZE)
}

func init() { ice.ChatCtxCmd(installed{}) }
