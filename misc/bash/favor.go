package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		FAVOR: {Name: FAVOR, Help: "收藏夹", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,type,name,text,pwd,username,hostname",
		)},
	}, Commands: map[string]*ice.Command{
		"/favor": {Name: "/favor", Help: "收藏", Action: map[string]*ice.Action{
			mdb.EXPORT: {Name: "export zone name", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				m.Echo("#!/bin/sh\n\n")
				m.Cmdy(FAVOR, m.Option(kit.MDB_ZONE)).Table(func(index int, value map[string]string, head []string) {
					if m.Option(kit.MDB_NAME) == "" || m.Option(kit.MDB_NAME) == value[kit.MDB_NAME] {
						m.Echo("# %v\n%v\n\n", value[kit.MDB_NAME], value[kit.MDB_TEXT])
					}
				})
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(FAVOR).Table()
		}},
		FAVOR: {Name: "favor zone id auto", Help: "收藏夹", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone=系统命令 type=shell name=1 text=pwd pwd=/home", Help: "添加"},
			cli.SYSTEM: {Name: "system", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, m.Option(cli.PWD))
				m.ProcessCommand(cli.SYSTEM, kit.Split(m.Option(kit.MDB_TEXT)), arg...)
				m.ProcessCommandOpt(arg, cli.PWD)
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Action(mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
			} else {
				m.PushAction(cli.SYSTEM)
				m.StatusTimeCount()
			}
		}},
	}})
}
