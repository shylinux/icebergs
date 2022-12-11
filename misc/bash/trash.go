package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const TRASH = "trash"

func init() {
	const (
		FROM = "from"
		TO   = "to"
	)
	Index.MergeCommands(ice.Commands{
		TRASH: {Name: "trash hash path auto", Help: "回收站", Actions: mdb.HashAction(mdb.FIELD, "time,hash,username,hostname,size,from,to")},
		web.PP(TRASH): {Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert from to", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, m.OptionSimple(aaa.USERNAME, tcp.HOSTNAME, nfs.SIZE, FROM, TO))
			}},
			mdb.REVERT: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m, m.Option(mdb.HASH))
				defer mdb.HashRemove(m, m.OptionSimple(mdb.HASH))
				m.Echo("mv %s %s", m.Append(TO), m.Append(FROM))
			}},
		}), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelectValue(m, func(key string, fields []string, value, val ice.Map) {
				kit.If(value[tcp.HOSTNAME] == m.Option(tcp.HOSTNAME), func() { m.Push(key, value, fields, val) })
			})
		}},
	})
}
