package chrome

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const SYNC = "sync"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SYNC: {Name: SYNC, Help: "同步流", Value: kit.Data(kit.MDB_FIELD, "time,id,type,name,text")},
	}, Commands: map[string]*ice.Command{
		"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(SYNC, mdb.INSERT, arg)
		}},
		SYNC: {Name: "sync id auto page export import", Help: "同步流", Action: ice.MergeAction(map[string]*ice.Action{
			FAVOR: {Name: "favor zone=hi name", Help: "收藏", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(FAVOR, mdb.INSERT)
			}},
		}, mdb.ListAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.ListSelect(m, arg...)
			m.PushAction(FAVOR)
		}},
	}})
}
