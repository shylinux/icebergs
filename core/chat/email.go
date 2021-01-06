package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _email_show(m *ice.Message, name, text string, arg ...string) {
}
func _email_search(m *ice.Message, kind, name, text string) {
	m.PushSearch("cmd", EMAIL, "type", "163", "name", name, "text", kit.MergeURL("https://mail.163.com"))
}

const EMAIL = "email"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		EMAIL: {Name: "email [name] word", Help: "百度", Action: map[string]*ice.Action{
			mdb.SEARCH: {Name: "search type name text", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				_email_search(m, arg[0], arg[1], arg[2])
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_email_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
		}},
	}})
}
