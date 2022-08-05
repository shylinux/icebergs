package tmux

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const SCRIPT = "script"

func init() {
	Index.MergeCommands(ice.Commands{
		SCRIPT: {Name: "script name auto create export import", Help: "脚本", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type=shell,tmux,vim name=hi text:textarea=pwd", Help: "添加"},
		}, mdb.HashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...)
		}},
	})
}
