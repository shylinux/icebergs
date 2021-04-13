package git

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const CONFIG = "config"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		CONFIG: {Name: "server name auto create", Help: "配置键", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create name value", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(cli.SYSTEM, GIT, CONFIG, "--global", m.Option(kit.MDB_NAME), m.Option(kit.MDB_VALUE))
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(cli.SYSTEM, GIT, CONFIG, "--global", arg)
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(cli.SYSTEM, GIT, CONFIG, "--global", "--unset", m.Option(kit.MDB_NAME))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			for _, v := range strings.Split(m.Cmdx(cli.SYSTEM, GIT, CONFIG, "--list"), "\n") {
				if ls := strings.Split(v, "="); len(ls) > 1 {
					m.Push(kit.MDB_NAME, ls[0])
					m.Push(kit.MDB_VALUE, ls[1])
					m.PushButton(mdb.REMOVE)
				}
			}
			m.Sort(kit.MDB_NAME)
		}},
	}})
}
