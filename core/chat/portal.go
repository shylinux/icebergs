package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
)

const PORTAL = "portal"

func init() {
	Index.MergeCommands(ice.Commands{
		PORTAL: {Name: "portal path auto", Help: "门户", Actions: ice.MergeActions(ice.Actions{
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) { web.RenderMain(m) }},
		}), Hand: func(m *ice.Message, arg ...string) { web.RenderMain(m) }},
	})
}
