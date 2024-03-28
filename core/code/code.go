package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CODE = "code"

var Index = &ice.Context{Name: CODE, Help: "编程中心"}

func init() {
	web.Index.Register(Index, &web.Frame{},
		WEBPACK, BINPACK, AUTOGEN, VERSION,
		COMPILE, PUBLISH, UPGRADE, INSTALL,
		XTERM, INNER, VIMER, BENCH, PPROF,
		TEMPLATE, COMPLETE, NAVIGATE,
		PACKAGE,
	)
}
func Prefix(arg ...ice.Any) string { return web.Prefix(CODE, kit.Keys(arg...)) }
