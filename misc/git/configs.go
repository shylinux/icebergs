package git

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _configs_set(m *ice.Message, key, value string) {
	m.Cmd(cli.SYSTEM, GIT, "config", "--global", key, value)
}
func _configs_get(m *ice.Message, key string) string {
	return m.Cmdx(cli.SYSTEM, GIT, "config", "--global", key)
}
func _configs_list(m *ice.Message) {
	for _, v := range strings.Split(_configs_get(m, "--list"), ice.NL) {
		if ls := strings.Split(v, "="); len(ls) > 1 {
			m.Push(mdb.NAME, ls[0])
			m.Push(mdb.VALUE, ls[1])
			m.PushButton(mdb.REMOVE)
		}
	}
	m.Sort(mdb.NAME)

	mdb.HashSelect(m.Spawn(ice.OptionFields("name,value"))).Table(func(index int, value map[string]string, head []string) {
		m.Push("", value, head).PushButton(mdb.CREATE)
	})
}

const CONFIGS = "configs"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		CONFIGS: {Name: CONFIGS, Help: "配置键", Value: kit.Data(
			mdb.SHORT, mdb.NAME, ice.INIT, kit.Dict(
				"alias", kit.Dict("s", "status", "b", "branch", "l", "log --oneline --decorate"),
				"credential", kit.Dict("helper", "store"),
				"core", kit.Dict("quotepath", "false"),
				"push", kit.Dict("default", "simple"),
				"color", kit.Dict("ui", "always"),
			))},
	}, Commands: map[string]*ice.Command{
		CONFIGS: {Name: "configs name auto create import", Help: "配置键", Action: map[string]*ice.Action{
			mdb.IMPORT: {Name: "import", Help: "初始化", Hand: func(m *ice.Message, arg ...string) {
				kit.Fetch(m.Configv(ice.INIT), func(conf string, value ice.Any) {
					kit.Fetch(value, func(key string, value string) { _configs_set(m, kit.Keys(conf, key), value) })
				})
			}},
			mdb.CREATE: {Name: "create name value", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.NAME))
				_configs_set(m, m.Option(mdb.NAME), m.Option(mdb.VALUE))
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.NAME, mdb.VALUE))
				_configs_set(m, "--unset", m.Option(mdb.NAME))
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.VALUE {
					m.Cmd(mdb.DELETE, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.NAME))
					_configs_set(m, m.Option(mdb.NAME), arg[1])
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_configs_list(m)
				return
			}
			m.Echo(_configs_get(m, arg[0]))
		}},
	}})
}
