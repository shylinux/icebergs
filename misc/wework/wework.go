package wework

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
)

const WEWORK = "wework"

var Index = &ice.Context{Name: WEWORK, Help: "企业微信"}

func init() { chat.Index.Register(Index, &web.Frame{}) }
