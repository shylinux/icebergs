package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CHAT = "chat"

var Index = &ice.Context{Name: CHAT, Help: "聊天中心"}

func init() {
	web.Index.Register(Index, &web.Frame{},
		HEADER, FOOTER,
		IFRAME, FAVOR,
		MESSAGE, TUTOR,
		FLOWS,
	)
}

func Prefix(arg ...string) string { return web.Prefix(CHAT, kit.Keys(arg)) }
