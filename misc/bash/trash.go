package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
)

const (
	FROM = "from"
	TO   = "to"
)
const TRASH = "trash"

func init() {
	Index.MergeCommands(ice.Commands{
		TRASH: {Name: "TRASH hash path auto prunes", Help: "回收站", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert from to", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME, nfs.SIZE, FROM, TO))
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				if !m.Warn(m.Option(TO) == "", ice.ErrNotValid, TO) {
					mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
					nfs.RemoveAll(m, m.Option(TO))
				}
			}},
			mdb.REVERT: {Name: "revert", Help: "恢复", Hand: func(m *ice.Message, arg ...string) {
				if !m.Warn(m.Option(FROM) == "" && m.Option(TO) == "", ice.ErrNotValid, FROM, TO) {
					nfs.Rename(m, m.Option(TO), m.Option(FROM))
					mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
				}
			}},
			mdb.PRUNES: {Name: "prunes before@date", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashPrunes(m, func(value ice.Map) bool {
					return true
				}).Tables(func(value ice.Maps) {
					nfs.RemoveAll(m, value[TO])
				})
			}},
			nfs.CAT: {Name: "cat", Help: "查看", Hand: func(m *ice.Message, arg ...string) {
				m.Option(nfs.DIR_ROOT, m.Option(TO))
				ctx.ProcessCommand(m, nfs.CAT, []string{}, arg...)
				ctx.ProcessCommandOpt(m, arg, TO)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,username,hostname,size,from,to"), web.ApiAction("/trash")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).PushAction(nfs.CAT, mdb.REVERT, mdb.REMOVE)
		}},
	})
}
