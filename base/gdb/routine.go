package gdb

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

const ROUTINE = "routine"

func init() {
	Index.MergeCommands(ice.Commands{
		ROUTINE: {Name: "routine hash auto prunes", Help: "协程池", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					cb := m.OptionCB("")
					h := mdb.HashCreate(m, m.OptionSimple(mdb.NAME), mdb.STATUS, START, ice.CMD, logs.FileLines(cb))
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
		}, mdb.HashStatusAction(mdb.FIELD, "time,hash,name,status,cmd"))},
	})
}
