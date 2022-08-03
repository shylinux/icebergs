package bash

import (
	"os"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
)

const (
	FROM = "from"
	TO   = "to"
)
const TRASH = "trash"

func init() {
	Index.MergeCommands(ice.Commands{
		"/trash": {Name: "/trash", Help: "回收", Actions: ice.Actions{
			mdb.INSERT: {Name: "insert from to", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(TRASH, mdb.INSERT, arg)
			}},
		}},
		TRASH: {Name: "TRASH hash path auto prunes", Help: "回收站", Actions: ice.MergeAction(ice.Actions{
			mdb.INSERT: {Name: "insert from to", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME, nfs.SIZE, FROM, TO))
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, "rm", "-rf", m.Option(TO))
				m.Cmdy(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH))
			}},
			mdb.REVERT: {Name: "revert", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, "mv", m.Option(TO), m.Option(FROM))
				m.Cmdy(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.HASH))
			}},
			mdb.PRUNES: {Name: "prunes before@date", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashPrunes(m, func(value ice.Maps) bool {
					os.RemoveAll(value[TO])
					return true
				})
			}},
			nfs.CAT: {Name: "cat", Help: "查看", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.DIR_ROOT, m.Option(TO))
				ctx.ProcessCommand(m, nfs.CAT, []string{}, arg...)
				ctx.ProcessCommandOpt(m, arg, TO)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,username,hostname,size,from,to")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...)
			m.PushAction(nfs.CAT, mdb.REVERT, mdb.REMOVE)
		}},
	})
}
