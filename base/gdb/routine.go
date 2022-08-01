package gdb

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

const ROUTINE = "routine"

func init() {
	Index.MergeCommands(ice.Commands{
		ROUTINE: {Name: "routine hash auto prunes", Help: "协程池", Actions: ice.MergeAction(ice.Actions{
			mdb.CREATE: {Name: "create name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					cb := m.OptionCB(ROUTINE)
					h := mdb.HashCreate(m, m.OptionSimple(mdb.NAME), mdb.STATUS, START, ice.CMD, logs.FileLine(cb, 100)).Result()
					defer func() {
						if e := recover(); e == nil {
							mdb.HashModify(m, mdb.HASH, h, mdb.STATUS, STOP)
						} else {
							mdb.HashModify(m, mdb.HASH, h, mdb.STATUS, ERROR, ERROR, e)
						}
					}()
					switch cb := cb.(type) {
					case []string:
						m.Cmd(kit.Split(kit.Join(cb)))
					case string:
						m.Cmd(kit.Split(cb))
					case func():
						cb()
					default:
						m.ErrorNotImplement(cb)
					}
				})
			}},
			"inner": {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				ls := kit.Split(m.Option(ice.CMD), ":")
				m.ProcessCommand("inner", []string{path.Dir(ls[0]), path.Base(ls[0]), ls[1]}, arg...)
			}},
		}, mdb.HashActionStatus(mdb.FIELD, "time,hash,name,status,cmd", mdb.ACTION, "inner"))},
	})
}
