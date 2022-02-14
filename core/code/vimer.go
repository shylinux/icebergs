package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const VIMER = "vimer"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		VIMER: {Name: "vimer path=src/ file=main.go line=1 auto", Help: "编辑器", Meta: kit.Dict(ice.DisplayLocal("", INNER)), Action: map[string]*ice.Action{
			nfs.SAVE: {Name: "save type file path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.SAVE, path.Join(m.Option(nfs.PATH), m.Option(nfs.FILE)))
			}},
			AUTOGEN: {Name: "create main=src/main.go zone name=hi help type=Zone,Hash,Lists,Data,Code list key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.CREATE, arg)
			}},
			COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				if msg := m.Cmd(COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN); !cli.IsSuccess(msg) {
					_inner_make(m, msg)
					return
				}
				m.Go(func() { m.Sleep("1s", ice.EXIT, "1") })
			}},
			BINPACK: {Name: "binpack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, BINPACK)
				m.Cmd(nfs.COPY, ice.GO_MOD, path.Join(ice.SRC_RELEASE, ice.GO_MOD))
				m.Cmd(nfs.COPY, ice.GO_SUM, path.Join(ice.SRC_RELEASE, ice.GO_SUM))
				m.Cmd(nfs.COPY, path.Join(ice.USR_RELEASE, "conf.go"), path.Join(ice.USR_ICEBERGS, "conf.go"))
				m.ProcessInner()
			}},
			DEVPACK: {Name: "devpack", Help: "开发模式", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.COPY, ice.GO_MOD, path.Join(ice.SRC_DEBUG, ice.GO_MOD))
				m.Cmd(nfs.COPY, ice.GO_SUM, path.Join(ice.SRC_DEBUG, ice.GO_SUM))
				m.Cmdy(WEBPACK, mdb.REMOVE)
				m.ToastSuccess()
				m.ProcessHold()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Cmdy(INNER, arg) }},
	}})
}
