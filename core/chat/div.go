package chat

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const DIV = "div"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		DIV: {Name: "div", Help: "定制", Value: kit.Data(
			kit.MDB_SHORT, "", kit.MDB_FIELD, "time,hash,type,name,text",
			kit.MDB_PATH, ice.USR_PUBLISH,
		)},
	}, Commands: map[string]*ice.Command{
		DIV: {Name: "div hash auto", Help: "定制", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.CREATE: {Name: "create type=page name=hi.html text", Help: "创建"},
			cli.MAKE: {Name: "make", Help: "生成", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(nfs.SAVE, path.Join(m.Conf(DIV, kit.META_PATH), m.Option(kit.MDB_NAME)), m.Option(kit.MDB_TEXT))
			}},
		}, mdb.HashAction(), mdb.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(arg), m.Conf(DIV, kit.META_FIELD))
			m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, kit.MDB_HASH, arg)
			m.Table(func(index int, value map[string]string, head []string) {
				m.PushAnchor("/" + path.Join(ice.PUBLISH, value[kit.MDB_NAME]))
			})
			if m.PushAction(cli.MAKE, mdb.REMOVE); len(arg) > 0 {
				m.Option(ice.MSG_DISPLAY, "/plugin/local/chat/div.js")
				m.Action("添加", "保存", "预览")
			} else {
				m.Action(mdb.CREATE)
			}
		}},
		"/div": {Name: "/div", Help: "定制", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.RenderIndex(web.SERVE, ice.VOLCANOS, "page/div.html")
		}},
	}})

}
