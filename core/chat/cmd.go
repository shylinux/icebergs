package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
)

const CMD = "cmd"

func init() {
	Index.MergeCommands(ice.Commands{
		CMD: {Help: "命令", Actions: web.ApiWhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			if len(arg[0]) == 0 || arg[0] == "" {
				web.RenderMain(m)
			} else {
				if m.IsCliUA() {
					m.Cmdy(arg, m.Optionv(ice.ARG)).RenderResult()
				} else {
					web.RenderCmd(m, arg[0], arg[1:])
				}
			}
		}},
	})
}
