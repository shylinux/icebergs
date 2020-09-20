package crx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
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
			SYNC: {Name: "sync id auto 导出 导入", Help: "同步流", Action: map[string]*ice.Action{
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
				FAVOR: {Name: "favor topic name", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(m.Prefix(FAVOR), mdb.INSERT, kit.MDB_TOPIC, m.Option(kit.MDB_TOPIC),
						kit.MDB_TYPE, SPIDE, kit.MDB_NAME, m.Option(kit.MDB_NAME), kit.MDB_TEXT, m.Option(kit.MDB_TEXT))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(mdb.FIELDS, kit.Select(m.Conf(SYNC, kit.META_FIELD), mdb.DETAIL, len(arg) > 0)); len(arg) > 0 {
					m.Option("cache.field", kit.MDB_ID)
					m.Option("cache.value", arg[0])
				} else {
					defer m.PushAction("收藏")
					if m.Option("_control", "page"); m.Option("cache.limit") == "" {
						m.Option("cache.limit", "10")
					}
				}

				m.Cmdy(mdb.SELECT, m.Prefix(SYNC), "", mdb.LIST, m.Option("cache.field"), m.Option("cache.value"))
			}},
			"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_TYPE, SPIDE,
					kit.MDB_NAME, m.Option(kit.MDB_NAME), kit.MDB_TEXT, m.Option(kit.MDB_TEXT))
			}},
			"/crx": {Name: "/crx", Help: "插件", Action: map[string]*ice.Action{
				web.HISTORY: {Name: "history", Help: "历史记录", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(web.SPIDE, web.SPIDE_SELF, "/code/chrome/sync",
						kit.MDB_NAME, arg[1], kit.MDB_TEXT, arg[2])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			}},
		},
	}, nil)
}
