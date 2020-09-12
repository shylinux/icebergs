package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
)

const CODE = "code"

var Index = &ice.Context{Name: CODE, Help: "编程中心",
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Conf(PUBLISH, "meta.contexts", _contexts)
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},
	},
}

func init() {
	web.Index.Register(Index, &web.Frame{},
		INSTALL, COMPILE, PUBLISH, UPGRADE,
		INNER, VIMER, BENCH, PPROF,
		C, SH, GO, SHY, JS,
	)
}
