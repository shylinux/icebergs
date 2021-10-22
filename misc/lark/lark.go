package lark

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
)

const LARK = "lark"

var Index = &ice.Context{Name: LARK, Help: "机器人"}

func init() { chat.Index.Register(Index, &web.Frame{}) }
