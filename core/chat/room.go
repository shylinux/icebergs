package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const ROOM = "room"
const JOIN = "join"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		ROOM: {Name: "room", Help: "room", Value: kit.Data(
			kit.MDB_SHORT, "zone",
			kit.MDB_FIELD, "time,id,type,name,text",
		)},
		JOIN: {Name: "join", Help: "join", Value: kit.Data(
			kit.MDB_SHORT, "zone",
			kit.MDB_FIELD, "time,hash,username,socket",
		)},
	}, Commands: map[string]*ice.Command{
		ROOM: {Name: "room zone id auto create insert join", Help: "room", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(ROOM), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
				m.Cmdy(mdb.INSERT, m.Prefix(JOIN), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
			}},
			mdb.INSERT: {Name: "insert zone type name text", Help: "发送", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(ROOM), "", mdb.ZONE, m.Option(kit.MDB_ZONE), arg[2:])
				m.Cmdy(JOIN, m.Option(kit.MDB_ZONE)).Table(func(index int, value map[string]string, head []string) {
					m.Option(ice.MSG_DAEMON, value[web.SOCKET])
					m.Toast(m.Option("text"), m.Option("name"))
				})
			}},

			"join": {Name: "join", Help: "加入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(JOIN, mdb.INSERT, arg)
			}},
			"exit": {Name: "exit", Help: "退出"},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(arg), mdb.ZONE_FIELD, m.Conf(ROOM, kit.META_FIELD))
			m.Cmdy(mdb.SELECT, m.Prefix(ROOM), "", mdb.ZONE, arg)
		}},
		JOIN: {Name: "join zone hash auto", Help: "join", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(ROOM), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
				m.Cmdy(mdb.INSERT, m.Prefix(JOIN), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
			}},
			mdb.INSERT: {Name: "insert zone username daemon", Hand: func(m *ice.Message, arg ...string) {
				m.Conf(m.Prefix(JOIN), kit.Keys(kit.KeyHash(m.Option(kit.MDB_ZONE)), kit.Keym(kit.MDB_SHORT)), web.SOCKET)
				m.Cmdy(mdb.INSERT, m.Prefix(JOIN), kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.HASH,
					aaa.USERNAME, m.Option(ice.MSG_USERNAME), web.SOCKET, m.Option(ice.MSG_DAEMON),
				)
			}},
			mdb.DELETE: {Name: "delete zone socket", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, m.Prefix(JOIN), kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.HASH, m.OptionSimple(web.SOCKET))
			}},
		}), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(arg), mdb.ZONE_FIELD, m.Conf(JOIN, kit.META_FIELD))
			if len(arg) == 0 {
				m.Cmdy(mdb.SELECT, m.Prefix(JOIN), "", mdb.HASH)
			} else {
				m.Cmdy(mdb.SELECT, m.Prefix(JOIN), kit.KeyHash(arg[0]), mdb.HASH, arg[1:])
				m.PushAction(mdb.REMOVE)
			}
		}},
	}})
}
