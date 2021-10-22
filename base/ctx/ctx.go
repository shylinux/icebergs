package ctx

import (
	ice "shylinux.com/x/icebergs"
)

const CTX = "ctx"

var Index = &ice.Context{Name: CTX, Help: "标准模块"}

func init() { ice.Index.Register(Index, nil, CONTEXT, COMMAND, CONFIG, MESSAGE) }
