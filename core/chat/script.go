package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
)

const SCRIPT = "script"

func init() {
	Index.MergeCommands(ice.Commands{
		SCRIPT: {Name: "script zone id auto", Help: "脚本化", Icon: "script.png", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone space index"},
		}, mdb.ImportantZoneAction(mdb.FIELDS, "time,id,space,index,play,status"),
		), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m.Options(mdb.CACHE_LIMIT, "30"), arg...).Display(""); len(arg) == 0 {
				m.PushAction(cli.RECORD, mdb.REMOVE).Action(mdb.CREATE, cli.STOP)
			} else {
				m.Sort(mdb.ID, ice.INT).PushAction("preview").Action(mdb.INSERT, cli.PLAY)
			}
		}},
	})
}
