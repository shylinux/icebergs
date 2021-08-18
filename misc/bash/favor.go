package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			FAVOR: {Name: FAVOR, Help: "收藏夹", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,type,name,text,pwd,username,hostname",
			)},
		},
		Commands: map[string]*ice.Command{
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
			FAVOR: {Name: "favor zone id auto", Help: "收藏夹", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert zone=系统命令 type=shell name=1 text=pwd pwd=/home", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
					m.Cmdy(mdb.INSERT, m.Prefix(FAVOR), kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.LIST, arg[2:])
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(FAVOR), "", mdb.ZONE, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(FAVOR), "", mdb.ZONE, m.OptionSimple(kit.MDB_ZONE))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(kit.MDB_ZONE, m.Conf(FAVOR, kit.META_FIELD))
					m.Cmdy(mdb.EXPORT, m.Prefix(FAVOR), "", mdb.ZONE)
					m.Conf(FAVOR, kit.MDB_HASH, "")
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(FAVOR), "", mdb.ZONE)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_ZONE:
						m.Cmdy(mdb.INPUTS, m.Prefix(FAVOR), "", mdb.HASH, arg)
					default:
						m.Cmdy(mdb.INPUTS, m.Prefix(FAVOR), kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.LIST, arg)
					}
				}},
				cli.SYSTEM: {Name: "system", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, m.Option(cli.PWD))
					m.ProcessCommand(cli.SYSTEM, kit.Split(m.Option(kit.MDB_TEXT)), arg...)
					m.ProcessCommandOpt(cli.PWD)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,zone,count", m.Conf(FAVOR, kit.META_FIELD))
				if m.Cmdy(mdb.SELECT, m.Prefix(FAVOR), "", mdb.ZONE, arg); len(arg) == 0 {
					m.Action(mdb.CREATE, mdb.EXPORT, mdb.IMPORT)
					m.PushAction(mdb.REMOVE)
				} else {
					m.PushAction(cli.SYSTEM)
				}
			}},
		},
	})
}
