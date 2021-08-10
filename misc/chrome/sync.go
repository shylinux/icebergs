package chrome

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const _sync_index = 1

func _sync_count(m *ice.Message) string { return m.Conf(SYNC, kit.Keym(kit.MDB_COUNT)) }

const SYNC = "sync"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SYNC: {Name: SYNC, Help: "同步流", Value: kit.Data(kit.MDB_FIELD, "time,id,type,name,text")},
	}, Commands: map[string]*ice.Command{
		"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(SYNC, mdb.INSERT, arg)
		}},
		SYNC: {Name: "sync id auto page export import", Help: "同步流", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case kit.MDB_ZONE:
					m.Cmdy(FAVOR, ice.OptionFields("zone,count,time"))
				}
			}},
			FAVOR: {Name: "favor zone=some name", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(FAVOR, mdb.INSERT, m.OptionSimple("zone,type,name,text"))
			}},
		}, mdb.ListAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.OptionPage(kit.Slice(arg, _sync_index)...)
			m.Fields(len(kit.Slice(arg, 0, 1)), m.Conf(SYNC, kit.META_FIELD))
			m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.LIST, kit.MDB_ID, arg)
			m.StatusTimeCountTotal(_sync_count(m))
			m.PushAction(FAVOR)
		}},
	}})
}
