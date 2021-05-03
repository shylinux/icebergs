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
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CONFIG: {Name: CONFIG, Help: "配置键", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			CONFIG: {Name: "server name auto create", Help: "配置键", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name value", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(cli.SYSTEM, GIT, CONFIG, "--global", m.Option(kit.MDB_NAME), m.Option(kit.MDB_VALUE))
					m.Cmd(mdb.DELETE, m.Prefix(CONFIG), "", kit.MDB_HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME))
					m.ProcessRefresh("0ms")
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					if arg[0] == kit.MDB_VALUE {
						m.Cmd(cli.SYSTEM, GIT, CONFIG, "--global", m.Option(kit.MDB_NAME), arg[1])
					}
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(cli.SYSTEM, GIT, CONFIG, "--global", "--unset", m.Option(kit.MDB_NAME))
					m.Cmd(mdb.INSERT, m.Prefix(CONFIG), "", kit.MDB_HASH, kit.MDB_NAME, m.Option(kit.MDB_NAME), kit.MDB_VALUE, m.Option(kit.MDB_VALUE))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					m.Echo(m.Cmdx(cli.SYSTEM, GIT, CONFIG, "--global", arg[0]))
					return
				}

				for _, v := range strings.Split(m.Cmdx(cli.SYSTEM, GIT, CONFIG, "--global", "--list"), "\n") {
					if ls := strings.Split(v, "="); len(ls) > 1 {
						m.Push(kit.MDB_NAME, ls[0])
						m.Push(kit.MDB_VALUE, ls[1])
						m.PushButton(mdb.REMOVE)
					}
				}
				m.Sort(kit.MDB_NAME)

				m.Cmd(mdb.SELECT, m.Prefix(CONFIG), "", kit.MDB_HASH, ice.OptionFields("name,value")).Table(func(index int, value map[string]string, head []string) {
					m.Push("", value, head)
					m.PushButton(mdb.CREATE)
				})
			}},
		}})
}
