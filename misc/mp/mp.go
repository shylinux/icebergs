package mp

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
)

const MP = "mp"

var Index = &ice.Context{Name: MP, Help: "小程序"}

func init() { chat.Index.Register(Index, &web.Frame{}) }
