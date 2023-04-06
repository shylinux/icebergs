package ctx

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const CTX = "ctx"

var Index = &ice.Context{Name: CTX, Help: "标准模块"}

func init() {
	ice.Index.Register(Index, nil, CONTEXT, COMMAND, CONFIG)
	ice.Info.Stack[Prefix()] = func(m *ice.Message, key string, arg ...ice.Any) ice.Any {
		return nil
	}
}
func Prefix(arg ...string) string { return kit.Keys(CTX, arg) }
