package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const CODE = "code"

var Index = &ice.Context{Name: CODE, Help: "编程中心", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Load()
		m.Conf(PUBLISH, kit.Keym("contexts"), _contexts)
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Save()
	}},
}}

func init() {
	web.Index.Register(Index, &web.Frame{},
		INSTALL, COMPILE, BINPACK, WEBPACK,
		VIMER, INNER, FAVOR, BENCH, PPROF,
		AUTOGEN, PUBLISH, UPGRADE,
		C, SH, SHY, GO, JS,
	)
}
