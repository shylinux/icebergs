package chrome

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const FIELD = "field"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		FIELD: {Name: "field", Help: "工具", Value: kit.Data(
			kit.MDB_SHORT, "zone", kit.MDB_FIELD, "time,id,index,args,style,left,top,selection",
		)},
	}, Commands: map[string]*ice.Command{
		FIELD: {Name: "field zone id auto insert", Help: "工具", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone=golang.google.cn index=cli.system args=pwd", Help: "添加"},
			SYNC: {Name: "sync hostname", Help: "同步", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(FIELD, m.Option("hostname")).Table(func(index int, value map[string]string, head []string) {
					m.Option(ice.MSG_OPTS, head)
					for k, v := range value {
						m.Option(k, v)
					}
					m.Cmd(web.SPACE, CHROME, CHROME, "1", m.Option("tid"), FIELD, value["index"], value["args"], value["top"])
				})
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Fields(len(arg), mdb.ZONE_FIELD, m.Conf(FIELD, kit.META_FIELD))
			if m.Cmdy(mdb.SELECT, m.PrefixKey(), "", mdb.ZONE, arg); len(arg) == 0 {
				m.PushAction(mdb.REMOVE)
			}
		}},
	}})
}
