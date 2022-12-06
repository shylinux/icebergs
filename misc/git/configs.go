package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _configs_set(m *ice.Message, key, value string) string {
	return _git_cmds(m, "config", "--global", key, value)
}
func _configs_get(m *ice.Message, key string) string {
	return _git_cmds(m, "config", "--global", key)
}
func _configs_list(m *ice.Message) {
	kit.SplitKV(ice.EQ, ice.NL, _configs_get(m, "--list"), func(text string, ls []string) {
		m.Push(mdb.NAME, ls[0]).Push(mdb.VALUE, ls[1]).PushButton(mdb.REMOVE)
	})
	mdb.HashSelectValue(m, func(value ice.Maps) { m.Push("", value, kit.Split("name,value")).PushButton(mdb.CREATE) })
	m.StatusTimeCount()
}

const CONFIGS = "configs"

func init() {
	Index.MergeCommands(ice.Commands{
		CONFIGS: {Name: "configs name value auto create import", Help: "配置键", Actions: ice.MergeActions(ice.Actions{
			mdb.IMPORT: {Help: "初始化", Hand: func(m *ice.Message, arg ...string) {
				kit.Fetch(m.Configv(ice.INIT), func(p string, v ice.Any) {
					kit.Fetch(v, func(k string, v string) { _configs_set(m, kit.Keys(p, k), v) })
				})
			}},
			mdb.CREATE: {Name: "create name* value*", Hand: func(m *ice.Message, arg ...string) {
				_configs_set(m, m.Option(mdb.NAME), m.Option(mdb.VALUE))
				mdb.HashRemove(m, m.Option(mdb.NAME))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m.Spawn(), m.OptionSimple(mdb.NAME, mdb.VALUE))
				_configs_set(m, "--unset", m.Option(mdb.NAME))
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.VALUE {
					mdb.HashRemove(m, m.Option(mdb.NAME))
					_configs_set(m, m.Option(mdb.NAME), arg[1])
				}
			}},
		}, mdb.HashAction(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,value", ice.INIT, kit.Dict(
				"alias", kit.Dict("s", "status", "b", "branch", "l", "log --oneline --decorate"),
				"credential", kit.Dict("helper", "store"),
				"core", kit.Dict("quotepath", "false"),
				"push", kit.Dict("default", "simple"),
				"color", kit.Dict("ui", "always"),
			),
		)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_configs_list(m)
			} else if len(arg) == 1 {
				m.Echo(_configs_get(m, arg[0]))
			} else {
				m.Echo(_configs_set(m, arg[0], arg[1]))
			}
		}},
	})
}
