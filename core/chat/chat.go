package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
)

const CHAT = "chat"

var Index = &ice.Context{Name: CHAT, Help: "聊天中心"}

func init() {
	web.Index.Register(Index, &web.Frame{},
		HEADER, RIVER, STORM, ACTION, FOOTER,
		SCAN, PASTE, FILES, LOCATION,
		WEBSITE,
	)
}
