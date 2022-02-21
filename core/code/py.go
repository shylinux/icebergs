package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const PY = "py"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		PY: {Name: "py", Help: "脚本", Value: kit.Data(INSTALL, kit.List(kit.Dict(
			cli.OSID, cli.ALPINE, ice.CMD, kit.List("apk", "add", "python2"),
		)))},
	}, Commands: map[string]*ice.Command{
		PY: {Name: "py", Help: "脚本", Action: map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.ENGINE, mdb.CREATE, PY, m.PrefixKey())
			}},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) {
				if !InstallSoftware(m.Spawn(), "python2", m.Configv(INSTALL)) {
					return
				}
				m.Cmdy(cli.SYSTEM, "python2", kit.Path(arg[2], arg[1]))
				m.Echo(ice.NL)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

		}},
	}})
}
