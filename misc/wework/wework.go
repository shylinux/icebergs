package wework

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
)

var Index = &ice.Context{Name: "wework", Help: "企业微信"}

func init() { chat.Index.Register(Index, &web.Frame{}) }
