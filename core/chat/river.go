package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _river_right(m *ice.Message, hash string) bool {
	if m.Option(ice.MSG_USERROLE) == aaa.ROOT {
		return true
	}
	return kit.IsIn(mdb.Conf(m, RIVER, kit.Keys(mdb.HASH, hash, mdb.META, mdb.TYPE)), "", aaa.VOID, m.Option(ice.MSG_USERROLE))
}
func _river_key(m *ice.Message, key ...ice.Any) string {
	return kit.Keys(mdb.HASH, m.Option(ice.MSG_RIVER), kit.Simple(key))
}
func _river_list(m *ice.Message) {
	if m.Option(web.SHARE) != "" {
		switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(mdb.TYPE) {
		case web.FIELD, web.STORM:
			m.Option(ice.MSG_RIVER, web.SHARE)
			return
		}
	}
	m.Cmd(mdb.SELECT, m.ShortKey(), "", mdb.HASH, ice.OptionFields(mdb.HASH, mdb.NAME, mdb.ICON, "main"), func(value ice.Maps) {
		kit.If(_river_right(m, value[mdb.HASH]), func() { m.PushRecord(value, mdb.HASH, mdb.NAME, mdb.ICON, "main") })
	})
	m.Sort(mdb.NAME)
}

const (
	RIVER_CREATE = "river.create"
)
const RIVER = "river"

func init() {
	Index.MergeCommands(ice.Commands{
		RIVER: {Help: "导航", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type=void,tech,root name=hi text=hello template=base", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m, arg)
				defer m.Result(h)
				kit.If(m.Option(mdb.TYPE) == aaa.VOID, func() { m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, kit.Keys(RIVER, h)) })
				gdb.Event(m, RIVER_CREATE, RIVER, m.Option(ice.MSG_RIVER, h), arg)
			}},
		}, web.ApiWhiteAction(), mdb.ImportantHashAction(mdb.FIELD, "time,hash,type,icon,name,text,template"), mdb.ExportHashAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.WarnNotLogin(m.Option(ice.MSG_USERNAME) == "") || !aaa.Right(m, RIVER, arg) {
				return
			} else if len(arg) == 0 {
				_river_list(m)
			} else if len(arg) > 1 && arg[1] == STORM {
				m.Cmdy(arg[1], arg[2:], kit.Dict(ice.MSG_RIVER, arg[0]))
			} else if len(arg) > 2 && arg[2] == STORM {
				m.Cmdy(arg[2], arg[3:], kit.Dict(ice.MSG_RIVER, arg[0], ice.MSG_STORM, arg[1]))
			}
		}},
	})
}
