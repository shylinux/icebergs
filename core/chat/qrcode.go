package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/core/wiki"
	kit "github.com/shylinux/toolkits"
)

const SCAN = "scan"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SCAN: {Name: SCAN, Help: "扫码", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			SCAN: {Name: "scan hash auto 添加@scan", Help: "扫码", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert text:textarea=hi", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Conf(SCAN, kit.Keys(m.Option(ice.MSG_DOMAIN), kit.MDB_META, kit.MDB_SHORT), kit.MDB_TEXT)
					m.Cmdy(mdb.INSERT, m.Prefix(SCAN), m.Option(ice.MSG_DOMAIN), mdb.HASH, arg)
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(SCAN), m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_TEXT, m.Option(kit.MDB_TEXT))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 0 {
					text := m.Cmd(mdb.SELECT, m.Prefix(SCAN), m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_HASH, arg[0]).Append(kit.MDB_TEXT)
					m.Cmdy(wiki.SPARK, "inner", text)
					m.Cmdy(wiki.IMAGE, "qrcode", text)
					m.Render("")
					return
				}

				m.Cmdy(mdb.SELECT, m.Prefix(SCAN), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				m.Sort(kit.MDB_TIME, "time_r")
				m.PushAction("删除")
			}},
		},
	}, nil)
}
