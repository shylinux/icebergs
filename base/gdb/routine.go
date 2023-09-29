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
		ROUTINE: {Help: "协程池", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create name cmd", Hand: func(m *ice.Message, arg ...string) {
				m.Go(func() {
					cb := m.OptionCB("")
					m.OptionDefault(ice.CMD, logs.FileLine(cb))
					h := mdb.HashCreate(m, m.OptionSimple(mdb.NAME, ice.CMD), mdb.STATUS, START)
					defer func() {
						if e := recover(); e == nil {
							mdb.HashRemove(m, mdb.HASH, h)
						} else {
							mdb.HashModify(m, mdb.HASH, h, mdb.STATUS, ERROR, ERROR, e)
						}
					}()
					switch cb := cb.(type) {
					case string:
						m.Cmd(kit.Split(cb))
					case []string:
						m.Cmd(kit.Split(kit.Join(cb)))
					case func(*ice.Message):
						cb(m.Spawn(m.Source()))
					case func():
						cb()
					default:
						m.ErrorNotImplement(cb)
					}
				}, m.Option(mdb.NAME))
			}},
		}, mdb.StatusHashAction(mdb.LIMIT, 1000, mdb.LEAST, 500, mdb.FIELD, "time,hash,status,name,cmd"), mdb.ClearOnExitHashAction())},
	})
}
func Go(m *ice.Message, cb ice.Any, arg ...string) {
	m.Cmd(ROUTINE, mdb.CREATE, kit.Select(m.PrefixKey(), arg, 0), logs.FileLine(cb), cb)
}
