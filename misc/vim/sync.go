package vim

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
)

const SYNC = "sync"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SYNC: {Name: SYNC, Help: "同步流", Value: kit.Data(
				kit.MDB_FIELD, "time,id,type,name,text",
			)},
		},
		Commands: map[string]*ice.Command{
			web.LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if f, _, e := m.R.FormFile("sub"); e == nil {
					defer f.Close()
					if b, e := ioutil.ReadAll(f); e == nil {
						m.Option("sub", string(b))
					}
				}
			}},

			SYNC: {Name: "sync id=auto auto 导出 导入", Help: "同步流", Action: map[string]*ice.Action{
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(SYNC), "", mdb.LIST)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(SYNC), "", mdb.LIST)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select(m.Conf(SYNC, kit.META_FIELD), mdb.DETAIL, len(arg) > 0))
				if len(arg) > 0 {
					m.Option("cache.field", kit.MDB_ID)
					m.Option("cache.value", arg[0])
				} else {
					if m.Option("_control", "page"); m.Option("cache.limit") == "" {
						m.Option("cache.limit", "10")
					}
				}

				m.Cmdy(mdb.SELECT, m.Prefix(SYNC), "", mdb.LIST, m.Option("cache.field"), m.Option("cache.value"))
			}},
			"/sync": {Name: "/sync", Help: "同步", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Render(ice.RENDER_RESULT)
				m.Cmd(mdb.INSERT, m.Prefix(SYNC), "", mdb.LIST, kit.MDB_TYPE, VIMRC, kit.MDB_NAME, arg[0], kit.MDB_TEXT, kit.Select(m.Option("arg"), m.Option("sub")),
					"pwd", m.Option("pwd"), "buf", m.Option("buf"), "row", m.Option("row"), "col", m.Option("col"))
			}},
		},
	})
}
