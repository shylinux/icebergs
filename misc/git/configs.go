package git

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _configs_set(m *ice.Message, k, v string) string { return _git_cmds(m, CONFIG, GLOBAL, k, v) }
func _configs_get(m *ice.Message, k string) string {
	return _git_cmds(m, CONFIG, GLOBAL, k)
}
func _configs_list(m *ice.Message) *ice.Message {
	if nfs.Exists(m, kit.HomePath(_GITCONFIG)) {
		kit.SplitKV(mdb.EQ, lex.NL, _configs_get(m, LIST), func(text string, ls []string) {
			m.Push(mdb.NAME, ls[0]).Push(mdb.VALUE, ls[1]).PushButton(mdb.REMOVE)
		})
	}
	return mdb.HashSelectValue(m, func(value ice.Maps) {
		m.Push("", value, kit.Split("name,value")).PushButton(mdb.CREATE)
	})
}

const (
	USER_EMAIL = "user.email"
	USER_NAME  = "user.name"

	GLOBAL = "--global"
	UNSET  = "--unset"
	LIST   = "--list"
)
const CONFIGS = "configs"

func init() {
	Index.MergeCommands(ice.Commands{
		CONFIGS: {Name: "configs name value auto", Help: "配置键", Actions: ice.MergeActions(ice.Actions{
			ice.INIT: {Help: "初始化", Hand: func(m *ice.Message, arg ...string) {
				kit.For(mdb.Configv(m, ice.INIT), func(p string, v ice.Any) {
					kit.For(v, func(k string, v string) { _configs_set(m, kit.Keys(p, k), v) })
				})
			}},
			mdb.CREATE: {Name: "create name* value*", Hand: func(m *ice.Message, arg ...string) {
				_configs_set(m, m.Option(mdb.NAME), m.Option(mdb.VALUE))
				mdb.HashRemove(m, m.Option(mdb.NAME))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashCreate(m, m.OptionSimple(mdb.NAME, mdb.VALUE))
				_configs_set(m, UNSET, m.Option(mdb.NAME))
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.VALUE {
					_configs_set(m, m.Option(mdb.NAME), arg[1])
					mdb.HashRemove(m, m.Option(mdb.NAME))
				}
			}},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,value", ice.INIT, kit.Dict(
			"alias", kit.Dict("s", STATUS, "b", BRANCH, "l", "log --oneline --decorate"),
			"push", kit.Dict("default", "simple"), "credential", kit.Dict("helper", "store"),
			"core", kit.Dict("quotepath", "false"), "color", kit.Dict("ui", "always"),
		))), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_configs_list(m).Action(mdb.CREATE, ice.INIT)
			} else {
				kit.If(len(arg) > 1, func() { _configs_set(m, arg[0], arg[1]) })
				m.Echo(_configs_get(m, arg[0]))
			}
		}},
	})
}
