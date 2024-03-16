package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CMD = "cmd"

func init() {
	Index.MergeCommands(ice.Commands{
		CMD: {Help: "命令", Actions: web.ApiWhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			switch cmd := kit.Select(web.WIKI_WORD, arg, 0); cmd {
			case web.ADMIN:
				web.RenderMain(m)
			default:
				if m.IsCliUA() {
					if aaa.Right(m, arg) {
						m.Cmdy(arg, m.Optionv(ice.ARG)).RenderResult()
					}
				} else {
					web.RenderCmd(m, cmd, kit.Slice(arg, 1))
				}
			}
		}},
	})
}
