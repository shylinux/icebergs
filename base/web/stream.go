package web

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const STREAM = "stream"

func init() {
	Index.MergeCommands(ice.Commands{
		STREAM: {Name: "stream hash daemon auto", Help: "在线", Actions: ice.MergeActions(ice.Actions{
			ONLINE: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, SPACE, m.Option(ice.MSG_SPACE), ctx.INDEX, m.Option(ice.MSG_INDEX),
					mdb.SHORT, cli.DAEMON, mdb.FIELD, mdb.Config(m, mdb.FIELDS))
				m.Option(mdb.SUBKEY, kit.Keys(mdb.HASH, kit.Hashs(kit.Fields(m.Option(ice.MSG_SPACE), m.Option(ice.MSG_INDEX)))))
				mdb.HashCreate(m, ParseUA(m))
				mdb.HashSelect(m.Options(ice.MSG_FIELDS, mdb.Config(m, mdb.FIELDS)))
			}},
			"push": {Hand: func(m *ice.Message, arg ...string) {
				m.Option(mdb.SUBKEY, kit.Keys(mdb.HASH, kit.Hashs(kit.Fields(m.Option(ice.MSG_SPACE), m.Option(ice.MSG_INDEX)))))
				mdb.HashSelect(m).Table(func(value ice.Maps) { m.Cmd(SPACE, value[cli.DAEMON], arg) })
			}},
		}, mdb.HashAction(
			mdb.SHORT, "space,index", mdb.FIELD, "time,hash,space,index",
			mdb.FIELDS, "time,daemon,userrole,username,usernick,avatar,icons,agent,system,ip,ua",
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				mdb.HashSelect(m)
			} else {
				m.Option(mdb.SUBKEY, kit.Keys(mdb.HASH, arg[0]))
				mdb.HashSelect(m.Options(ice.MSG_FIELDS, mdb.Config(m, mdb.FIELDS)), arg[1:]...)
			}
		}},
	})
}
