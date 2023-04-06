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
		INSTALL, UPGRADE, WEBPACK, BINPACK, AUTOGEN, COMPILE, PUBLISH,
		VIMER, INNER, XTERM, PPROF, BENCH,
		C, SH, SHY, PY, GO, JS, CSS, HTML,
	)
	ice.Info.Stack[CODE] = func(m *ice.Message, key string, arg ...ice.Any) ice.Any {
		return nil
	}
}

func Prefix(arg ...string) string { return web.Prefix(CODE, kit.Keys(arg)) }
