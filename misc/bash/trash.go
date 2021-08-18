package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
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
				kit.MDB_FIELD, "time,hash,username,hostname,size,from,to",
			)},
		},
		Commands: map[string]*ice.Command{
			"/trash": {Name: "/trash", Help: "回收", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert from to", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(TRASH, mdb.INSERT, arg)
				}},
			}},
			TRASH: {Name: "TRASH hash path auto prunes", Help: "回收站", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert from= to=", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(TRASH), "", mdb.HASH, m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME, kit.MDB_SIZE, FROM, TO))
				}},
				mdb.REVERT: {Name: "revert", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "mv", m.Option(TO), m.Option(FROM))
					m.Cmdy(mdb.DELETE, m.Prefix(TRASH), "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "rm", "-rf", m.Option(TO))
					m.Cmdy(mdb.DELETE, m.Prefix(TRASH), "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				}},
				nfs.DIR: {Name: "dir", Help: "目录", Hand: func(m *ice.Message, arg ...string) {
					m.Option(nfs.DIR_ROOT, m.Option(TO))
					m.ProcessCommand(nfs.DIR, []string{}, arg...)
					m.ProcessCommandOpt(TO)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), m.Conf(TRASH, kit.META_FIELD))
				m.Cmdy(mdb.SELECT, m.Prefix(TRASH), "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(nfs.DIR, mdb.REVERT, mdb.REMOVE)
			}},
		},
	})
}
