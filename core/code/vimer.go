package code

import (
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
		VIMER: {Name: "vimer path=src/ file=main.go line=1 刷新:button=auto save", Help: "编辑器", Meta: kit.Dict(
			ice.Display("/plugin/local/code/vimer.js", "editor"),
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
			AUTOGEN: {Name: "create main=src/main.go@key key= zone= type=Zone,Hash,List,Data name=hi list= help=", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.CREATE, arg)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(AUTOGEN, mdb.INPUTS, arg)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(INNER, arg)
		}},
	}})
}
