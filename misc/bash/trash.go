package bash

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
)

const (
	FROM = "from"
	TO   = "to"
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
					m.Cmdy(mdb.INSERT, m.Prefix(TRASH), "", mdb.HASH, FROM, m.Option(FROM), TO, m.Option(TO))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "rm", "-rf", m.Option(TO))
					m.Cmdy(mdb.DELETE, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				}},
				"reback": {Name: "reback", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "mv", m.Option(TO), m.Option(FROM))
					m.Cmdy(mdb.DELETE, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				nfs.DIR: {Name: "dir", Help: "目录", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) > 0 && arg[0] == cli.RUN {
						m.Option(nfs.DIR_ROOT, m.Option(TO))
						m.Cmdy(nfs.DIR, kit.Select("", arg, 1))
						return
					}
					m.ShowPlugin("", nfs.DIR, cli.RUN)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select(m.Conf(m.Prefix(TRASH), kit.META_FIELD), mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(nfs.DIR, "reback", mdb.REMOVE)
			}},

			"/trash": {Name: "/trash", Help: "回收", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert from to", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(TRASH), "", mdb.HASH, tcp.HOSTNAME, m.Option(tcp.HOSTNAME),
						kit.MDB_SIZE, m.Option(kit.MDB_SIZE), FROM, m.Option(FROM), TO, m.Option(TO))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			}},
		},
	})
}
