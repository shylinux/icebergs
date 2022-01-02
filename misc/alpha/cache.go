package alpha

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const CACHE = "cache"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CACHE: {Name: "cache", Help: "缓存", Value: kit.Data(
			mdb.SHORT, "word", mdb.FIELD, "time,word,translation,definition",
		)},
	}, Commands: map[string]*ice.Command{
		CACHE: {Name: "cache word auto", Help: "缓存", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create word translation definition", Help: "创建"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	}})
}
