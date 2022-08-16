package git

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _configs_set(m *ice.Message, key, value string) string {
	return _git_cmds(m, "config", "--global", key, value)
}
func _configs_get(m *ice.Message, key string) string {
	if msg := _git_cmd(m, "config", "--global", key); cli.IsSuccess(msg) {
		return msg.Result()
	}
	return ""
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

	mdb.HashSelectValue(m, func(value ice.Maps) {
		m.Push("", value, kit.Split("name,value")).PushButton(mdb.CREATE)
	})
}

const CONFIGS = "configs"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		CONFIGS: {Name: CONFIGS, Help: "配置键", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,value", ice.INIT, kit.Dict(
				"alias", kit.Dict("s", "status", "b", "branch", "l", "log --oneline --decorate"),
				"credential", kit.Dict("helper", "store"),
				"core", kit.Dict("quotepath", "false"),
				"push", kit.Dict("default", "simple"),
				"color", kit.Dict("ui", "always"),
			))},
	}, Commands: ice.Commands{
		CONFIGS: {Name: "configs name auto create import", Help: "配置键", Actions: ice.Actions{
			mdb.IMPORT: {Name: "import", Help: "初始化", Hand: func(m *ice.Message, arg ...string) {
				kit.Fetch(m.Configv(ice.INIT), func(conf string, value ice.Any) {
					kit.Fetch(value, func(key string, value string) { _configs_set(m, kit.Keys(conf, key), value) })
				})
			}},
			mdb.CREATE: {Name: "create name value", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				_configs_set(m, m.Option(mdb.NAME), m.Option(mdb.VALUE))
				mdb.HashRemove(m, m.Option(mdb.NAME))
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m.Spawn(), m.OptionSimple(mdb.NAME, mdb.VALUE))
				_configs_set(m, "--unset", m.Option(mdb.NAME))
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.VALUE {
					mdb.HashRemove(m, m.Option(mdb.NAME))
					_configs_set(m, m.Option(mdb.NAME), arg[1])
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_configs_list(m)
			} else {
				m.Echo(_configs_get(m, arg[0]))
			}
		}},
	}})
}
