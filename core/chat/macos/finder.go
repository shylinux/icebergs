package macos

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
)

const FINDER = "finder"

func init() {
	Index.MergeCommands(ice.Commands{
		FINDER: {Name: "finder list", Help: "应用", Role: aaa.VOID, Actions: ice.MergeActions(ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(APPLICATIONS, mdb.INPUTS, arg) }},
			mdb.INSERT: {Name: "insert space index* args name* icon*@icon", Hand: func(m *ice.Message, arg ...string) { m.Cmdy(APPLICATIONS, mdb.CREATE, arg) }},
		}, CmdHashAction(mdb.NAME))},
	})
}

func FinderAppend(m *ice.Message, icon, index string) { install(m, FINDER, icon, index) }
