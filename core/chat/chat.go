package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/web"
)

const CHAT = "chat"

var Index = &ice.Context{Name: CHAT, Help: "聊天中心",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(web.SERVE, aaa.WHITE, HEADER, RIVER, ACTION, FOOTER)
			m.Load()
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save()
		}},
	},
}

func init() {
	web.Index.Register(Index, &web.Frame{},
		HEADER, RIVER, STORM, ACTION, FOOTER,
		SCAN, PASTE, FILES, LOCATION,
	)
}
