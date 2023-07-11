package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const FINDER = "finder"

func init() {
	Index.MergeCommands(ice.Commands{FINDER: {Name: "finder list", Actions: CmdHashAction(mdb.NAME)}})
}

func FinderAppend(m *ice.Message, name, index string, arg ...string) {
	install(m, FINDER, name, index, arg...)
}
