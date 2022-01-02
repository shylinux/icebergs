package tmux

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const SCRIPT = "script"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		SCRIPT: {Name: SCRIPT, Help: "脚本", Value: kit.Data(
			mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text",
		)},
	}, Commands: map[string]*ice.Command{
		SCRIPT: {Name: "script name auto create export import", Help: "脚本", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create type=shell,tmux,vim name=hi text:textarea=pwd", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	}})
}
