package code

import ice "shylinux.com/x/icebergs"

const PORTAL = "portal"

func init() {
	Index.MergeCommands(ice.Commands{
		PORTAL: {Name: "portal path auto", Help: "门户", Hand: func(m *ice.Message, arg ...string) {
		}},
	})
}
