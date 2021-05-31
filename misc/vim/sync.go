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
			SYNC: {Name: "sync id auto", Help: "同步流", Action: map[string]*ice.Action{
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Option(mdb.FIELDS, mdb.DETAIL)
				} else {
					m.Option(mdb.FIELDS, m.Conf(SYNC, kit.META_FIELD))
				}

				m.Cmdy(mdb.SELECT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_ID, arg)
			}},
			"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch m.Option(ARG) {
				case "wq", "q", "qa":
					m.Cmd("/sess", LOGOUT)
				}

				m.Cmd(mdb.INSERT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_TYPE, VIMRC,
					kit.MDB_NAME, arg[0], kit.MDB_TEXT, kit.Select(m.Option(ARG), m.Option(SUB)),
					PWD, m.Option(PWD), BUF, m.Option(BUF), ROW, m.Option(ROW), COL, m.Option(COL))
			}},
		},
	})
}
