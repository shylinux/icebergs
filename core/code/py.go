package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const PY = "py"

func init() {
	const PYTHON2 = "python2"
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		PY: {Name: "py", Help: "脚本", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.ENGINE, mdb.CREATE, PY, m.PrefixKey())
				m.Cmd(mdb.RENDER, mdb.CREATE, PY, m.PrefixKey())
			}},
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) {
				if kit.FileExists(kit.Path(arg[2], arg[1])) {
					m.Cmdy(cli.SYSTEM, PYTHON2, kit.Path(arg[2], arg[1]))
				} else if b, ok := ice.Info.Pack[path.Join(arg[2], arg[1])]; ok && len(b) > 0 {
					m.Cmdy(cli.SYSTEM, PYTHON2, "-c", string(b))
				}
				if cli.IsSuccess(m) {
					m.SetAppend()
				}
				m.Echo(ice.NL)
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				if kit.FileExists(kit.Path(arg[2], arg[1])) {
					m.Cmdy(cli.SYSTEM, PYTHON2, kit.Path(arg[2], arg[1]))
				} else if b, ok := ice.Info.Pack[path.Join(arg[2], arg[1])]; ok && len(b) > 0 {
					m.Cmdy(cli.SYSTEM, PYTHON2, "-c", string(b))
				}
				m.Echo(ice.NL)
			}},
		}},
	}})
}
