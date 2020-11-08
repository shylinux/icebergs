package bash

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"path"
)

const TRASH = "trash"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TRASH: {Name: TRASH, Help: "回收站", Value: kit.Data(
				kit.MDB_FIELD, "time,hash,hostname,size,from,to",
			)},
		},
		Commands: map[string]*ice.Command{
			TRASH: {Name: "TRASH hash path auto prunes", Help: "回收站", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert from= to=", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(TRASH), "", mdb.HASH, "from", m.Option("from"), "to", m.Option("to"))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "rm", "-rf", m.Option("to"))
					m.Cmdy(mdb.DELETE, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				"reback": {Name: "reback", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "mv", m.Option("to"), m.Option("from"))
					m.Cmdy(mdb.DELETE, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Cmd(mdb.SELECT, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, arg).Table(func(index int, value map[string]string, head []string) {
						m.Cmdy(nfs.DIR, path.Join(value["to"], kit.Select("", arg, 1)))
					})
					return
				}
				m.Option(mdb.FIELDS, kit.Select(m.Conf(m.Prefix(TRASH), kit.META_FIELD), mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction("reback", mdb.REMOVE)
			}},

			"/trash": {Name: "/trash", Help: "回收", Action: map[string]*ice.Action{
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(TRASH), "", mdb.HASH, tcp.HOSTNAME, m.Option(tcp.HOSTNAME),
					kit.MDB_SIZE, m.Option(kit.MDB_SIZE), "from", m.Option("from"), "to", m.Option("to"))
			}},
		},
	})
}
