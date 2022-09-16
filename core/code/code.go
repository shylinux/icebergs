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
		INSTALL, WEBPACK, BINPACK, AUTOGEN, COMPILE, PUBLISH, UPGRADE,
		XTERM, VIMER, INNER, FAVOR, BENCH, PPROF,
		C, SH, SHY, GO, JS,
	)
}

func Prefix(arg ...string) string { return kit.Keys(web.WEB, CODE, arg) }
