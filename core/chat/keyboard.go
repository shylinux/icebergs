package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const KEYBOARD = "keyboard"

func init() {
	Index.MergeCommands(ice.Commands{
		KEYBOARD: {Name: "keyboard hash@keyboard auto", Help: "键盘", Actions: ice.MergeActions(ice.Actions{
			"_refresh": {Name: "refresh", Help: "刷新", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPACE, m.Option("space"), "refresh")
			}},
			"inputs": {Name: "refresh", Help: "刷新", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPACE, m.Option("space"), "refresh")
			}},
		}, mdb.HashAction(mdb.SHORT, "", mdb.FIELD, "time,hash,space,index,input")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Cmd(web.SPACE, m.Option("space"), arg)
				return
			}
			mdb.HashSelect(m, arg...)
			if len(arg) > 0 && arg[0] != "" {
				meta := kit.UnMarshal(m.Cmd(ctx.COMMAND, m.Append("index")).Append("meta"))
				list := []string{}
				kit.Fetch(meta, func(key string, value ice.Any) {
					if key == "_trans" {
						return
					}
					list = append(list, key)
				})
				m.PushAction(kit.Join(list))
				m.Option("meta", kit.Format(meta))
			}
		}},
	})
}
