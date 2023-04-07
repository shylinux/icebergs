package ctx

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const CTX = "ctx"

var Index = &ice.Context{Name: CTX, Help: "标准模块"}

func init() { ice.Index.Register(Index, nil, CONTEXT, COMMAND, CONFIG) }

func Prefix(arg ...string) string { return kit.Keys(CTX, arg) }
