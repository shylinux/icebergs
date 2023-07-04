package wiki

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const STYLE = "style"

func init() {
	Index.MergeCommands(ice.Commands{
		STYLE: {Name: "style class auto", Help: "样式", Hand: func(m *ice.Message, arg ...string) {
			switch kit.Select("end", arg, 0) {
			case "end":
				m.Echo("</div>")
			default:
				if len(arg) > 1 {
					m.Echo(`<div class="%s %s" style="%s">`, "story", arg[0], kit.JoinKV(":", ";", arg[1:]...))
				} else {
					m.Echo(`<div class="%s %s">`, "story", arg[0])
				}
			}
		}},
	})
}
