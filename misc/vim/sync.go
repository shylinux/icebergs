package vim

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const _sync_index = 1

func _sync_count(m *ice.Message) string {
	return m.Conf(SYNC, kit.Keym(kit.MDB_COUNT))
}

const SYNC = "sync"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SYNC: {Name: SYNC, Help: "同步流", Value: kit.Data(
				kit.MDB_FIELD, "time,id,type,name,text,pwd,username,hostname",
			)},
		},
		Commands: map[string]*ice.Command{
			"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				switch m.Option(ARG) {
				case "wq", "q", "qa":
					m.Cmd("/sess", aaa.LOGOUT)
				}

				m.Cmd(mdb.INSERT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_TYPE, VIMRC,
					kit.MDB_NAME, arg[0], kit.MDB_TEXT, kit.Select(m.Option(ARG), m.Option(SUB)),
					cli.PWD, m.Option(cli.PWD), BUF, m.Option(BUF), ROW, m.Option(ROW), COL, m.Option(COL))
			}},
			SYNC: {Name: "sync id auto page", Help: "同步流", Action: map[string]*ice.Action{
				mdb.PREV: {Name: "prev", Help: "上一页", Hand: func(m *ice.Message, arg ...string) {
					mdb.PrevPage(m, _sync_count(m), kit.Slice(arg, _sync_index)...)
				}},
				mdb.NEXT: {Name: "next", Help: "下一页", Hand: func(m *ice.Message, arg ...string) {
					mdb.NextPage(m, _sync_count(m), kit.Slice(arg, _sync_index)...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.OptionPage(kit.Slice(arg, _sync_index)...)
				m.Fields(len(kit.Slice(arg, 0, 1)), m.Conf(SYNC, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_ID, arg)
				m.StatusTimeCountTotal(_sync_count(m))
			}},
		},
	})
}
