package tmux

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const SCRIPT = "script"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		SCRIPT: {Name: SCRIPT, Help: "脚本", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text",
		)},
	}, Commands: ice.Commands{
		SCRIPT: {Name: "script name auto create export import", Help: "脚本", Actions: ice.MergeAction(ice.Actions{
			mdb.CREATE: {Name: "create type=shell,tmux,vim name=hi text:textarea=pwd", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	}})
}
