package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const ROOM = "room"

func init() {
	const (
		JOIN = "join"
		QUIT = "quit"
	)
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		ROOM: {Name: "room", Help: "room", Value: kit.Data(
			kit.MDB_SHORT, "zone", kit.MDB_FIELD, "time,id,type,name,text",
		)},
		JOIN: {Name: "join", Help: "join", Value: kit.Data(
			kit.MDB_SHORT, "space", kit.MDB_FIELD, "time,hash,username,socket",
		)},
	}, Commands: map[string]*ice.Command{
		ROOM: {Name: "room zone id auto", Help: "room", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(ROOM), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
			}},
			mdb.INSERT: {Name: "insert zone type=hi name=hello text=world", Help: "发送", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(ROOM), "", mdb.ZONE, m.Option(kit.MDB_ZONE), arg[2:])
				m.Cmdy(mdb.SELECT, m.Prefix(ROOM), kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.HASH, ice.Option{mdb.FIELDS, "time,space"}).Table(func(index int, value map[string]string, head []string) {
					m.Debug("what %v", value)
					m.Cmdy(web.SPACE, value[web.SPACE], "toast", m.Option("text"), m.Option("name"))
				})
			}},

			JOIN: {Name: "join zone", Help: "加入", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(ROOM), kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.HASH, web.SPACE, m.Option("_daemon"))
				m.Cmdy(mdb.INSERT, m.Prefix(JOIN), "", mdb.HASH, web.SPACE, m.Option("_daemon"), kit.MDB_SHORT, kit.MDB_ZONE)
				m.Cmdy(mdb.INSERT, m.Prefix(JOIN), kit.KeyHash(m.Option("_daemon")), mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
			}},
			QUIT: {Name: "quit", Help: "退出", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, m.Prefix(ROOM), kit.KeyHash(m.Option(kit.MDB_ZONE)), mdb.HASH, m.OptionSimple(web.SPACE))
				m.Cmdy(mdb.DELETE, m.Prefix(JOIN), kit.KeyHash(m.Option(web.SPACE)), mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if mdb.ZoneSelect(m, arg...); len(arg) == 0 {
				m.Action(mdb.CREATE)
				m.PushAction(mdb.INSERT, JOIN)
			} else if len(arg) == 1 {
				m.Action(mdb.INSERT, JOIN)
			}
		}},
		JOIN: {Name: "join space zone auto", Help: "join", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				// m.Watch(web.SPACE_START, m.PrefixKey())
			}},
			web.SPACE_START: {Name: "space_start", Help: "下线", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, m.Prefix(ROOM), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
				m.Cmdy(mdb.INSERT, m.Prefix(JOIN), "", mdb.HASH, m.OptionSimple(kit.MDB_ZONE))
			}},
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
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(arg), "time,space", "time,zone")
			if len(arg) == 0 {
				m.Cmdy(mdb.SELECT, m.Prefix(JOIN), "", mdb.HASH)
			} else {
				m.Cmdy(mdb.SELECT, m.Prefix(JOIN), kit.KeyHash(arg[0]), mdb.HASH, arg[1:])
				m.PushAction(mdb.REMOVE)
			}
		}},
	}})
}
