package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _river_right(m *ice.Message, hash string) bool {
	return m.Option(ice.MSG_USERROLE) == aaa.ROOT || m.Cmdx(aaa.ROLE, aaa.RIGHT, m.Option(ice.MSG_USERROLE), RIVER, hash) == ice.OK
}
func _river_key(m *ice.Message, key ...ice.Any) string {
	return kit.Keys(mdb.HASH, m.Option(ice.MSG_RIVER), kit.Simple(key))
}
func _river_list(m *ice.Message) {
	if m.Option(web.SHARE) != "" {
		switch msg := m.Cmd(web.SHARE, m.Option(web.SHARE)); msg.Append(mdb.TYPE) {
		case web.STORM, web.FIELD:
			m.Option(ice.MSG_RIVER, "_share")
			return
		}
	}
	m.Cmd(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, ice.OptionFields(mdb.HASH, mdb.NAME), func(value ice.Maps) {
		if _river_right(m, value[mdb.HASH]) {
			m.PushRecord(value, mdb.HASH, mdb.NAME)
		}
	})
}

const (
	RIVER_CREATE = "river.create"
)
const RIVER = "river"

func init() {
	Index.MergeCommands(ice.Commands{
		web.P(RIVER): {Name: "/river hash auto create", Help: "群组", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(aaa.ROLE, aaa.BLACK, aaa.VOID, m.CommandKey(), ctx.ACTION)
			}},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case nfs.TEMPLATE:
					m.Cmdy(TEMPLATE).CutTo(RIVER, arg[0])
				default:
					mdb.HashInputs(m, arg)
				}
			}},
			mdb.CREATE: {Name: "create type=void,tech name=hi text=hello template=base", Hand: func(m *ice.Message, arg ...string) {
				h := mdb.HashCreate(m, arg)
				defer m.Result(h)

				if m.Option(mdb.TYPE) == aaa.VOID {
					m.Cmd(aaa.ROLE, aaa.WHITE, aaa.VOID, kit.Keys(RIVER, h))
				}
				gdb.Event(m, RIVER_CREATE, RIVER, m.Option(ice.MSG_RIVER, h), arg)
			}},
			RIVER_CREATE: {Name: "river.create", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.EXPORT, m.Prefix(RIVER), "", mdb.HASH)
				m.Cmd(mdb.IMPORT, m.Prefix(RIVER), "", mdb.HASH)
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,template"), aaa.RoleAction()), Hand: func(m *ice.Message, arg ...string) {
			if m.Warn(m.Option(ice.MSG_USERNAME) == "", ice.ErrNotLogin) {
				m.RenderStatusUnauthorized()
				return
			}
			if !aaa.Right(m, RIVER, arg) {
				m.RenderStatusForbidden()
				return
			}
			if len(arg) == 0 {
				_river_list(m)
				return
			}
			if len(arg) > 1 && arg[1] == STORM {
				m.Option(ice.MSG_RIVER, arg[0])
				m.Cmdy(arg[1], arg[2:])
				return
			}
			if len(arg) > 2 && arg[2] == STORM {
				m.Option(ice.MSG_RIVER, arg[0])
				m.Option(ice.MSG_STORM, arg[1])
				m.Cmdy(arg[2], arg[3:])
				return
			}
		}},
	})
}
