package wiki

import (
	ice "github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const WIKI = "wiki"

var Index = &ice.Context{Name: WIKI, Help: "文档中心",
	Configs: map[string]*ice.Config{
		WIKI: {Name: WIKI, Help: "文档中心", Value: kit.Data(
			kit.MDB_FIELD, "time,hash,type,name,text",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		WIKI: {Name: WIKI, Help: "文档中心", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(WIKI), "", mdb.HASH, arg)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(mdb.SELECT, m.Prefix(WIKI), "", mdb.HASH, kit.MDB_HASH, arg)
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
