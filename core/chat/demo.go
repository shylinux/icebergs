package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
)

func init() {
	Index.MergeCommands(ice.Commands{
		"demo-hash": {Actions: ice.MergeActions(mdb.HashAction(), ctx.CmdAction())},
		"demo-status-hash": {Actions: ice.MergeActions(mdb.StatusHashAction(), ctx.CmdAction())},
		"demo-list": {Actions: ice.MergeActions(mdb.ListAction(), ctx.CmdAction())},
		"demo-page-list": {Actions: ice.MergeActions(mdb.PageListAction(), ctx.CmdAction())},
		"demo-zone": {Actions: ice.MergeActions(mdb.ZoneAction(), ctx.CmdAction())},
	})
}
