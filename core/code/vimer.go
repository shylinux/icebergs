package code

import (
	"os"
	"path"
	"strings"

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
			ice.Display("/plugin/local/code/vimer.js", "inner"),
		), Action: map[string]*ice.Action{
			mdb.ENGINE: {Name: "engine", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmdy(mdb.ENGINE, arg); len(m.Resultv()) > 0 || len(m.Appendv(ice.MSG_APPEND)) > 0 {
					return
				}

				if arg = kit.Split(strings.Join(arg, " ")); !m.Warn(!m.Right(arg)) {
					if m.Cmdy(arg); len(m.Appendv(ice.MSG_APPEND)) == 0 && len(m.Resultv()) == 0 {
						m.Cmdy(cli.SYSTEM, arg)
					}
				}
			}},
			nfs.SAVE: {Name: "save type file path", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.SAVE, path.Join(m.Option(kit.MDB_PATH), m.Option(kit.MDB_FILE)))
			}},
			AUTOGEN: {Name: "create main=src/main.go@key key zone type=Zone,Hash,Data name=hi list help", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.CREATE, arg)
			}},
			COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_ENV,
					cli.HOME, os.Getenv(cli.HOME), cli.PATH, os.Getenv(cli.PATH),
					"CGO_ENABLED", "0", "GOCACHE", os.Getenv("GOCACHE"),
					"GOPRIVATE", "shylinux.com",
				)

				_autogen_version(m)
				if m.Cmdy(cli.SYSTEM, "go", "build", "-v", "-o", "bin/ice.bin", "src/main.go", "src/version.go"); m.Append(cli.CMD_CODE) == "0" {
					m.Cmd("exit", "1")
				}
				m.ProcessInner()
			}},
			BINPACK: {Name: "binpack", Help: "打包", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, BINPACK)
				p := path.Join("src/release", ice.GO_MOD)
				if _, e := os.Stat(p); e == nil {
					m.Cmd(nfs.COPY, ice.GO_MOD, p)
				}
				m.ProcessInner()
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.INPUTS, arg)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(INNER, arg)
		}},
	}})
}
