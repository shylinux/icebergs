package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _email_show(m *ice.Message, name, text string, arg ...string) {
}
func _email_search(m *ice.Message, kind, name, text string) {
	if kit.Contains(kind, kit.MDB_FOREACH) || kit.Contains(kind, EMAIL) {
		m.PushSearchWeb(EMAIL, name)
	}
}

const EMAIL = "email"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			EMAIL: {Name: "email", Help: "邮件", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			EMAIL: {Name: "email [name] word", Help: "邮件", Action: map[string]*ice.Action{
				mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
					_email_search(m, arg[0], arg[1], arg[2])
				}},
				mdb.CREATE: {Name: "create type name text", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(mdb.INSERT, m.Prefix(EMAIL), "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_email_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
		}})
}
