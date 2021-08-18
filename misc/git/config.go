package git

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _config_set(m *ice.Message, key, value string) {
	m.Cmd(cli.SYSTEM, GIT, CONFIG, "--global", key, value)
}
func _config_get(m *ice.Message, key string) string {
	return m.Cmdx(cli.SYSTEM, GIT, CONFIG, "--global", key)
}
func _config_list(m *ice.Message) {
	for _, v := range strings.Split(_config_get(m, "--list"), "\n") {
		if ls := strings.Split(v, "="); len(ls) > 1 {
			m.Push(kit.MDB_NAME, ls[0])
			m.Push(kit.MDB_VALUE, ls[1])
			m.PushButton(mdb.REMOVE)
		}
	}
	m.Sort(kit.MDB_NAME)

	m.Cmd(mdb.SELECT, m.Prefix(CONFIG), "", mdb.HASH, ice.OptionFields("name,value")).Table(func(index int, value map[string]string, head []string) {
		m.Push("", value, head)
		m.PushButton(mdb.CREATE)
	})
}

const CONFIG = "config"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CONFIG: {Name: CONFIG, Help: "配置键", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_NAME, "init", kit.Dict(
				"alias", kit.Dict("s", "status", "b", "branch"),
				"credential", kit.Dict("helper", "store"),
				"core", kit.Dict("quotepath", "false"),
				"push", kit.Dict("default", "simple"),
				"color", kit.Dict("ui", "always"),
			))},
	}, Commands: map[string]*ice.Command{
		CONFIG: {Name: "server name auto create import", Help: "配置键", Action: map[string]*ice.Action{
			mdb.IMPORT: {Name: "import", Help: "初始化", Hand: func(m *ice.Message, arg ...string) {
				kit.Fetch(m.Confv(CONFIG, kit.Keym("init")), func(conf string, value interface{}) {
					kit.Fetch(value, func(key string, value string) {
						_config_set(m, kit.Keys(conf, key), value)
					})
				})
			}},
			mdb.CREATE: {Name: "create name value", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.DELETE, m.Prefix(CONFIG), "", mdb.HASH, m.OptionSimple(kit.MDB_NAME))
				_config_set(m, m.Option(kit.MDB_NAME), m.Option(kit.MDB_VALUE))
				m.ProcessRefresh30ms()
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == kit.MDB_VALUE {
					m.Cmd(mdb.DELETE, m.Prefix(CONFIG), "", mdb.HASH, m.OptionSimple(kit.MDB_NAME))
					_config_set(m, m.Option(kit.MDB_NAME), arg[1])
					m.ProcessRefresh30ms()
				}
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.Prefix(CONFIG), "", mdb.HASH, m.OptionSimple(kit.MDB_NAME, kit.MDB_VALUE))
				_config_set(m, "--unset", m.Option(kit.MDB_NAME))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				_config_list(m)
				return
			}
			m.Echo(_config_get(m, arg[0]))
		}},
	}})
}
