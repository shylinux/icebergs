package gdb

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const ROUTINE = "routine"

func init() {
	Index.Merge(&ice.Context{Configs: ice.Configs{
		ROUTINE: {Name: ROUTINE, Help: "协程池", Value: kit.Data(mdb.SHORT, "time,hash,status,fileline")},
	}, Commands: ice.Commands{
		ROUTINE: {Name: "routine hash auto prunes", Help: "协程池", Actions: ice.MergeAction(ice.Actions{
			mdb.CREATE: {Name: "create fileline status", Help: "创建"},
			mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
				m.OptionFields(m.Config(mdb.SHORT))
				m.Cmdy(mdb.PRUNES, ROUTINE, "", mdb.HASH, cli.STATUS, cli.STOP)
				m.Cmdy(mdb.PRUNES, ROUTINE, "", mdb.HASH, cli.STATUS, cli.ERROR)
			}},
			"inner": {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				ls := kit.Split(m.Option("fileline"), ":")
				m.ProcessCommand("inner", []string{path.Dir(ls[0]), path.Base(ls[0]), ls[1]}, arg...)
			}},
		}, mdb.HashAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...)
			m.PushAction("inner", mdb.REMOVE)
		}},
	}})
}
