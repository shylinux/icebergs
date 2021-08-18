package bash

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

const _sync_index = 1

func _sync_count(m *ice.Message) string {
	return m.Conf(SYNC, kit.Keym(kit.MDB_COUNT))
}

const (
	SHELL   = "shell"
	HISTORY = "history"
)
const SYNC = "sync"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SYNC: {Name: SYNC, Help: "同步流", Value: kit.Data(
				kit.MDB_FIELD, "time,id,type,name,text,pwd,username,hostname",
			)},
		},
		Commands: map[string]*ice.Command{
			"/sync": {Name: "/sync", Help: "同步", Action: map[string]*ice.Action{
				HISTORY: {Name: "history", Help: "历史", Hand: func(m *ice.Message, arg ...string) {
					ls := strings.SplitN(strings.TrimSpace(m.Option(ARG)), " ", 4)
					if text := strings.TrimSpace(strings.Join(ls[3:], " ")); text != "" {
						m.Cmd(mdb.INSERT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_TIME, ls[1]+" "+ls[2],
							kit.MDB_TYPE, SHELL, kit.MDB_NAME, ls[0], kit.MDB_TEXT, text,
							m.OptionSimple(cli.PWD, aaa.USERNAME, tcp.HOSTNAME, tcp.HOSTNAME))

					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.INSERT, m.Prefix(SYNC), "", mdb.HASH, kit.MDB_TYPE, SHELL, kit.MDB_NAME, arg[0],
					kit.MDB_TEXT, m.Option(SUB), m.OptionSimple(cli.PWD))
			}},
			SYNC: {Name: "sync id auto page export import", Help: "同步流", Action: map[string]*ice.Action{
				mdb.PREV: {Name: "prev", Help: "上一页", Hand: func(m *ice.Message, arg ...string) {
					mdb.PrevPage(m, _sync_count(m), kit.Slice(arg, _sync_index)...)
				}},
				mdb.NEXT: {Name: "next", Help: "下一页", Hand: func(m *ice.Message, arg ...string) {
					mdb.NextPage(m, _sync_count(m), kit.Slice(arg, _sync_index)...)
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.OptionFields(m.Conf(SYNC, kit.META_FIELD))
					m.Cmdy(mdb.EXPORT, m.Prefix(SYNC), "", mdb.LIST)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(SYNC), "", mdb.LIST)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_ZONE:
						m.Cmdy(FAVOR, mdb.INPUTS, arg)
					}
				}},
				cli.SYSTEM: {Name: "system", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, m.Option(cli.PWD))
					m.ProcessCommand(cli.SYSTEM, kit.Split(m.Option(kit.MDB_TEXT)), arg...)
					m.ProcessCommandOpt(cli.PWD)
				}},
				FAVOR: {Name: "favor zone=some@key type name text pwd", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(FAVOR, mdb.INSERT, m.OptionSimple(kit.MDB_ZONE, m.Conf(FAVOR, kit.META_FIELD)))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.OptionPage(kit.Slice(arg, _sync_index)...)
				m.Fields(len(kit.Slice(arg, 0, 1)), m.Conf(SYNC, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_ID, arg)
				m.PushAction(cli.SYSTEM, FAVOR)
				m.StatusTimeCountTotal(_sync_count(m))
			}},
		},
	})
}
