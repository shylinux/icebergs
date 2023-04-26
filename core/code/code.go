package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/yac"
	kit "shylinux.com/x/toolkits"
)

const CODE = "code"

var Index = &ice.Context{Name: CODE, Help: "编程中心"}

func init() {
	web.Index.Register(Index, &web.Frame{},
		INSTALL, UPGRADE, WEBPACK, BINPACK, AUTOGEN, COMPILE, PUBLISH,
		VIMER, INNER, XTERM, PPROF, BENCH,
		C, SH, SHY, PY, GO, JS, CSS, HTML,
		TEMPLATE, COMPLETE, NAVIGATE,
	)
}
func init() {
	return
	ice.Info.Stack[Prefix(Index.Register)] = func(m *ice.Message, key string, arg ...ice.Any) ice.Any {
		return Index.Register(yac.TransContext(m, Prefix(), arg...), &web.Frame{})
	}
}
func Prefix(arg ...ice.Any) string { return web.Prefix(CODE, kit.Keys(arg...)) }
