package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/ssh"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const AUTH = "auth"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		AUTH: {Name: "auth hash auto create", Help: "授权", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create type=node,user name text", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m, AUTH), mdb.HASH,
					aaa.USERROLE, m.Option(ice.MSG_USERROLE), aaa.USERNAME, m.Option(ice.MSG_USERNAME),
					kit.MDB_TIME, m.Time("72h"), arg)
			}},
			mdb.INSERT: {Name: "insert river share", Help: "加入", Hand: func(m *ice.Message, arg ...string) {
				switch msg := m.Cmd(AUTH, m.Option(web.SHARE)); msg.Append(kit.MDB_TYPE) {
				case USER:
					m.Option(ice.MSG_RIVER, m.Option(RIVER))
					m.Cmdy(USER, mdb.INSERT, aaa.USERNAME, m.Option(ice.MSG_USERNAME))
				}
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.MODIFY, RIVER, _river_key(m, AUTH), mdb.HASH, m.OptionSimple(kit.MDB_HASH), arg)
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, RIVER, _river_key(m, AUTH), mdb.HASH, m.OptionSimple(kit.MDB_HASH))
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(arg), "time,hash,userrole,username,type,name,text")
			if m.Cmdy(mdb.SELECT, RIVER, _river_key(m, AUTH), mdb.HASH, kit.MDB_HASH, arg); len(arg) > 0 {
				m.PushScript(ssh.SCRIPT, _river_url(m, web.SHARE, m.Option(web.SHARE)))
				m.PushQRCode(cli.QRCODE, _river_url(m, web.SHARE, m.Option(web.SHARE)))
			}
			m.PushAction(mdb.REMOVE)
		}},
	}})
}
