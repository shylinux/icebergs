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
				m.Cmd(web.SPACE, m.Option(web.SPACE), "refresh")
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {}},
			"input": {Name: "input", Help: "刷新", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPACE, m.Option(web.SPACE), "input", arg)
				ctx.ProcessHold(m)
			}},
		}, mdb.HashAction(mdb.SHORT, "", mdb.FIELD, "time,hash,space,index,input")), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && arg[0] == ctx.ACTION {
				m.Cmd(web.SPACE, m.Option(web.SPACE), arg)
				ctx.ProcessHold(m)
				return
			}

			if mdb.HashSelect(m, arg...); len(arg) > 0 && arg[0] != "" {
				msg := m.Cmd(ctx.COMMAND, m.Append(ctx.INDEX))
				meta := kit.UnMarshal(msg.Append(mdb.META))
				list := kit.UnMarshal(msg.Append(mdb.LIST))
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
				m.Option(mdb.META, kit.Format(meta))
				ctx.DisplayLocal(m, "")
			}
		}},
	})
}
