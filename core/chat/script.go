package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
)

const SCRIPT = "script"

func init() {
	Index.MergeCommands(ice.Commands{
		SCRIPT: {Name: "script zone id auto", Help: "脚本化", Icon: "script.png", Actions: ice.MergeActions(
			mdb.ImportantZoneAction(mdb.FIELDS, "time,id,space,index,play,status,style"), mdb.ExportZoneAction(),
		), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m.Options(mdb.CACHE_LIMIT, "30"), arg...).Display(""); len(arg) == 0 {
				m.PushAction(cli.RECORD, mdb.REMOVE).Action(mdb.CREATE, cli.STOP)
			} else {
				m.Sort(mdb.ID, ice.INT).PushAction(web.PREVIEW).Action(mdb.INSERT, cli.PLAY)
			}
		}},
	})
}
