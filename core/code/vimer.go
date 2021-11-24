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
		VIMER: {Name: "vimer path=src/ file=main.go line=1 refresh:button=auto save", Help: "编辑器", Meta: kit.Dict(
			ice.Display("/plugin/local/code/vimer.js", INNER),
		), Action: map[string]*ice.Action{
			nfs.SAVE: {Name: "save type file path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.SAVE, path.Join(m.Option(kit.MDB_PATH), m.Option(kit.MDB_FILE)))
			}},
			ice.RUN: {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, GO, ice.RUN, path.Join(kit.Slice(arg, 0, 2)...))
				m.Set(ice.MSG_APPEND)
				m.ProcessInner()
			}},
			mdb.ENGINE: {Name: "engine", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmdy(mdb.ENGINE, arg); len(m.Resultv()) > 0 || m.Length() > 0 {
					return
				}

				if arg = kit.Split(kit.Join(arg, ice.SP)); m.Right(arg) {
					if m.Cmdy(arg); len(m.Resultv()) == 0 && m.Length() == 0 {
						m.Cmdy(cli.SYSTEM, arg)
					}
				}
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.INPUTS, arg)
			}},
			AUTOGEN: {Name: "create main=src/main.go@key key zone type=Zone,Hash,Data name=hi list help", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.CREATE, arg)
			}},
			COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmdy(COMPILE, ice.SRC_MAIN_GO); cli.IsSuccess(m) {
					m.Cmd(COMPILE, ice.SRC_MAIN_GO, ice.BIN_ICE_BIN)
					m.Cmd(ice.EXIT, "1")
				}
				m.ProcessInner()
			}},
			BINPACK: {Name: "binpack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, BINPACK)
				m.Cmd(nfs.COPY, ice.GO_MOD, path.Join(ice.SRC_RELEASE, ice.GO_MOD))
				m.Cmd(nfs.COPY, ice.GO_SUM, path.Join(ice.SRC_RELEASE, ice.GO_SUM))
				m.Cmd(nfs.COPY, path.Join(ice.USR_RELEASE, "conf.go"), path.Join(ice.USR_ICEBERGS, "conf.go"))
				m.ProcessInner()
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(INNER, arg)
		}},
	}})
}
