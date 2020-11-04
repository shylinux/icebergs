package vim

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const SYNC = "sync"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SYNC: {Name: SYNC, Help: "同步流", Value: kit.Data(
				kit.MDB_FIELD, "time,id,type,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			SYNC: {Name: "sync id=auto auto 导出 导入", Help: "同步流", Action: map[string]*ice.Action{
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(SYNC), "", mdb.LIST)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(SYNC), "", mdb.LIST)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_TOPIC:
						m.Cmdy(m.Prefix(FAVOR)).Appendv(ice.MSG_APPEND, kit.MDB_TOPIC, kit.MDB_COUNT, kit.MDB_TIME)
					}
				}},
				FAVOR: {Name: "favor topic type name text", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(m.Prefix(FAVOR), mdb.INSERT, kit.MDB_TOPIC, m.Option(kit.MDB_TOPIC),
						kit.MDB_TYPE, m.Option(kit.MDB_TYPE), kit.MDB_NAME, m.Option(kit.MDB_NAME), kit.MDB_TEXT, m.Option(kit.MDB_TEXT))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Option(mdb.FIELDS, mdb.DETAIL)
					m.Option(mdb.CACHE_FILED, kit.MDB_ID)
					m.Option(mdb.CACHE_VALUE, arg[0])
				} else {
					m.Option(mdb.FIELDS, m.Conf(SYNC, kit.META_FIELD))
					m.Option(ice.MSG_CONTROL, ice.CONTROL_PAGE)
					defer m.PushAction(FAVOR)
				}

				m.Cmdy(mdb.SELECT, m.Prefix(SYNC), "", mdb.LIST, m.Option(mdb.CACHE_FILED), m.Option(mdb.CACHE_VALUE))
			}},
			"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(ARG) == "qa" {
					m.Cmdy(mdb.MODIFY, m.Prefix(SESS), "", mdb.HASH, kit.MDB_HASH, m.Option(SID), kit.MDB_STATUS, "logout")
				}

				m.Cmd(mdb.INSERT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_TYPE, VIMRC, kit.MDB_NAME, arg[0],
					kit.MDB_TEXT, kit.Select(m.Option(ARG), m.Option(SUB)),
					PWD, m.Option(PWD), BUF, m.Option(BUF), ROW, m.Option(ROW), COL, m.Option(COL))
			}},
		},
	})
}
