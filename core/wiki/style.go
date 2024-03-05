package wiki

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const (
	COLUMN = "column"
	FLEX   = "flex"
	END    = "end"
)
const STYLE = "style"

func init() {
	Index.MergeCommands(ice.Commands{
		STYLE: {Name: "style class", Help: "样式", Hand: func(m *ice.Message, arg ...string) {
			switch kit.Select("end", arg, 0) {
			case "end":
				m.Echo("</div>")
			default:
				if len(arg) > 1 {
					m.Echo(`<div class="story %s" style="%s">`, arg[0], kit.JoinKV(":", ";", kit.TransArgKey(arg[1:], transKey)...))
				} else {
					m.Echo(`<div class="story %s">`, arg[0])
				}
			}
		}},
	})
}
