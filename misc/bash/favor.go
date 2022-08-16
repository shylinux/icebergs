package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		"/favor": {Name: "/favor", Help: "收藏", Actions: ice.Actions{
			mdb.EXPORT: {Name: "export zone name", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				m.Echo("#!/bin/sh\n\n")
				m.Cmdy(FAVOR, m.Option(mdb.ZONE), func(value ice.Maps) {
					if m.Option(mdb.NAME) == "" || m.Option(mdb.NAME) == value[mdb.NAME] {
						m.Echo("# %v\n%v\n\n", value[mdb.NAME], value[mdb.TEXT])
					}
				})
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			m.Cmdy(FAVOR).Table()
		}},
		FAVOR: {Name: "favor zone id auto", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone=系统命令 type=shell name=1 text=pwd pwd=/home", Help: "添加"},
			cli.SYSTEM: {Name: "system", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, m.Option(cli.PWD))
				ctx.ProcessCommand(m, cli.SYSTEM, kit.Split(m.Option(mdb.TEXT)), arg...)
				ctx.ProcessCommandOpt(m, arg, cli.PWD)
			}},
		}, mdb.ZoneAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,type,name,text,pwd,username,hostname")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
			} else {
				m.PushAction(cli.SYSTEM)
				m.StatusTimeCount()
			}
		}},
	})
}
