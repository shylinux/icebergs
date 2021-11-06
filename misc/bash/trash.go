package bash

import (
	"os"

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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		TRASH: {Name: TRASH, Help: "回收站", Value: kit.Data(
			kit.MDB_FIELD, "time,hash,username,hostname,size,from,to",
		)},
	}, Commands: map[string]*ice.Command{
		"/trash": {Name: "/trash", Help: "回收", Action: map[string]*ice.Action{
			mdb.INSERT: {Name: "insert from to", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TRASH, mdb.INSERT, arg)
			}},
		}},
		TRASH: {Name: "TRASH hash path auto prunes", Help: "回收站", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {Name: "insert from to", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME, kit.MDB_SIZE, FROM, TO))
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, "rm", "-rf", m.Option(TO))
				m.Cmdy(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
			}},
			mdb.REVERT: {Name: "revert", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, "mv", m.Option(TO), m.Option(FROM))
				m.Cmdy(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(kit.MDB_HASH))
			}},
			mdb.PRUNES: {Name: "prunes before@date", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashPrunes(m, func(value map[string]string) bool {
					os.RemoveAll(value[TO])
					return false
				})
			}},
			nfs.CAT: {Name: "cat", Help: "查看", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.DIR_ROOT, m.Option(TO))
				m.ProcessCommand(nfs.CAT, []string{}, arg...)
				m.ProcessCommandOpt(arg, TO)
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
			m.PushAction(nfs.CAT, mdb.REVERT, mdb.REMOVE)
		}},
	}})
}
