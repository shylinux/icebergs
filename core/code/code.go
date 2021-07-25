package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
)

const CODE = "code"

var Index = &ice.Context{Name: CODE, Help: "编程中心", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Load()
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Save()
	}},
}}

func init() {
	web.Index.Register(Index, &web.Frame{},
		WEBPACK, BINPACK, AUTOGEN, COMPILE, UPGRADE, PUBLISH, INSTALL,
		VIMER, INNER, FAVOR, BENCH, PPROF,
		C, SH, SHY, GO, JS,
	)
}
