package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CMD = "cmd"

func init() {
	Index.MergeCommands(ice.Commands{
		CMD: {Help: "命令", Actions: web.ApiWhiteAction(), Hand: func(m *ice.Message, arg ...string) {
			web.RenderCmd(m, kit.Select(web.WIKI_WORD, arg, 0), kit.Slice(arg, 1))
		}},
	})
}
