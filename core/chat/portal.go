package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
)

const PORTAL = "portal"

func init() {
	Index.MergeCommands(ice.Commands{
		PORTAL: {Help: "门户", Actions: web.ApiAction(), Hand: func(m *ice.Message, arg ...string) { web.RenderMain(m) }},
	})
}
