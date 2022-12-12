package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor zone id auto insert", Help: "收藏夹", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone*=demo type=shell name=1 text=pwd pwd=/home"},
			cli.SYSTEM: {Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] == ice.RUN {
					if msg := mdb.ZoneSelect(m.Spawn(), m.Option(mdb.ZONE), m.Option(mdb.ID)); nfs.ExistsFile(m, msg.Append(cli.PWD)) {
						m.Option(cli.CMD_DIR, msg.Append(cli.PWD))
					}
					ctx.ProcessField(m, "", nil, arg...)
				} else {
					ctx.ProcessField(m, "", kit.Split(m.Option(mdb.TEXT)))
				}
			}},
			web.DOWNLOAD: {Hand: func(m *ice.Message, arg ...string) { web.RenderCache(m, m.Option(mdb.TEXT)) }},
		}, mdb.ZoneAction(mdb.FIELD, "time,id,type,name,text,pwd,username,hostname")), Hand: func(m *ice.Message, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Action(mdb.EXPORT, mdb.IMPORT)
			} else {
				m.PushAction(kit.Select(cli.SYSTEM, web.DOWNLOAD, arg[0] == _DOWNLOAD))
			}
		}},
		web.PP(FAVOR): {Actions: ice.Actions{
			mdb.EXPORT: {Name: "export zone* name", Hand: func(m *ice.Message, arg ...string) {
				m.Echo("#!/bin/sh\n\n").Cmdy(FAVOR, m.Option(mdb.ZONE), func(value ice.Maps) {
					if m.Option(mdb.NAME) == "" || m.Option(mdb.NAME) == value[mdb.NAME] {
						m.Echo("# %s\n%s\n\n", value[mdb.NAME], value[mdb.TEXT])
					}
				})
			}},
		}, Hand: func(m *ice.Message, arg ...string) { m.Cmdy(FAVOR).Table() }},
	})
}
