package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const CODE = "code"

var Index = &ice.Context{Name: CODE, Help: "编程中心", Commands: ice.Commands{
	ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
		m.Conf(TEMPLATE, kit.Keym(mdb.SHORT), mdb.TYPE)
		m.Conf(COMPLETE, kit.Keym(mdb.SHORT), mdb.TYPE)
		m.Conf(NAVIGATE, kit.Keym(mdb.SHORT), mdb.TYPE)
		ctx.Load(m)
	}},
}}

func init() {
	web.Index.Register(Index, &web.Frame{},
		INSTALL, UPGRADE, WEBPACK, BINPACK, AUTOGEN, COMPILE, PUBLISH,
		FAVOR, XTERM, INNER, VIMER, PPROF, BENCH,
		C, SH, SHY, GO, JS,
	)
}

func Prefix(arg ...string) string { return kit.Keys(web.WEB, CODE, arg) }
