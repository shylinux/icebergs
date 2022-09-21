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
			}},
			"input": {Name: "input", Help: "刷新", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPACE, m.Option("space"), "input", arg)
				ctx.ProcessHold(m)
			}},
		}, mdb.HashAction(mdb.SHORT, "", mdb.FIELD, "time,hash,space,index,input")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Cmd(web.SPACE, m.Option("space"), arg)
				ctx.ProcessHold(m)
				return
			}
			mdb.HashSelect(m, arg...)
			if len(arg) > 0 && arg[0] != "" {
				msg := m.Cmd(ctx.COMMAND, m.Append("index"))
				meta := kit.UnMarshal(msg.Append("meta"))
				list := kit.UnMarshal(msg.Append("list"))
				keys := []string{}
				kit.Fetch(list, func(index int, value ice.Any) {
					if kit.Format(kit.Value(value, mdb.TYPE)) == "button" {
						return
					}
					keys = append(keys, kit.Format(kit.Value(value, mdb.NAME)))
				})
				kit.Fetch(meta, func(key string, value ice.Any) {
					if key == "_trans" {
						return
					}
					keys = append(keys, key)
				})
				m.PushAction(kit.Join(keys))
				m.Option("meta", kit.Format(meta))
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}
