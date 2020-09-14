package zsh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"
)

const TRASH = "trash"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TRASH: {Name: TRASH, Help: "收藏夹", Value: kit.Data(
				kit.MDB_FIELD, "time,hash,hostname,size,from,to",
			)},
		},
		Commands: map[string]*ice.Command{
			TRASH: {Name: "TRASH hash auto 清理", Help: "收藏夹", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert from= to=", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(TRASH), "", mdb.HASH, "from", m.Option("from"), "to", m.Option("to"))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "rm", "-rf", m.Option("to"))
					m.Cmdy(mdb.DELETE, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				}},
				"reback": {Name: "reback", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "mv", m.Option("to"), m.Option("from"))
					m.Cmdy(mdb.DELETE, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Cmd(mdb.SELECT, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, arg).Table(func(index int, value map[string]string, head []string) {
						m.Cmdy(nfs.DIR, value["to"])
					})
					return
				}
				m.Option(mdb.FIELDS, kit.Select(m.Conf(m.Prefix(TRASH), kit.META_FIELD), mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction("恢复", "删除")
			}},

			"/trash": {Name: "/trash", Help: "回收", Action: map[string]*ice.Action{
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(TRASH), "", mdb.HASH, "hostname", m.Option("hostname"),
					"size", m.Option("size"), "from", m.Option("from"), "to", m.Option("to"))
			}},
		},
	}, nil)
}
