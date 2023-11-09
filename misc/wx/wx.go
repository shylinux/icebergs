package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
)

const WX = "wx"

var Index = &ice.Context{Name: WX, Help: "微信公众号"}

func init() { chat.Index.Register(Index, &web.Frame{}) }
