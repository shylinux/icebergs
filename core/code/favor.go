package code

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor zone id auto insert page", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone*=数据结构 type=go name*=hi text*=hello path file line"},
			XTERM: {Help: "终端", Hand: func(m *ice.Message, arg ...string) {
				if msg := mdb.ZoneSelects(m.Spawn(), m.Option(mdb.ZONE), m.Option(mdb.ID)); strings.HasPrefix(msg.Option(mdb.TYPE), cli.OPEN) {
					m.Cmdy(cli.SYSTEM, m.Option(mdb.TYPE)).ProcessHold()
				} else {
					ctx.Process(m, m.ActionKey(), msg.OptionSimple(mdb.TYPE, mdb.NAME, mdb.TEXT), arg...)
				}
			}},
			INNER: {Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				msg := mdb.ZoneSelects(m, m.Option(mdb.ZONE), m.Option(mdb.ID))
				ctx.Process(m, m.ActionKey(), msg.OptionSplit(nfs.PATH, nfs.FILE, nfs.LINE), arg...)
			}},
		}, mdb.PageZoneAction(mdb.FIELD, "time,id,type,name,text,path,file,line"), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			m.Option(mdb.CACHE_LIMIT, "30")
			if mdb.PageZoneSelect(m, arg...); len(arg) > 0 && arg[0] != "" {
				m.Tables(func(value ice.Maps) { m.PushButton(kit.Select(INNER, XTERM, value[nfs.FILE] == "")) }).Option(ctx.STYLE, arg[0])
			}
		}},
	})
}
