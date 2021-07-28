package lark

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/chat"
	kit "github.com/shylinux/toolkits"
)

const HOME = "home"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		HOME: {Name: "home river storm title content", Help: "首页", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			name := kit.Select(m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_USERNICK))
			if len(name) > 10 {
				name = name[:10]
			}
			name += "的" + kit.Select("应用列表", arg, 2)

			text, link, list := kit.Select("", arg, 3), kit.MergeURL2(m.Conf(web.SHARE, kit.Keym(kit.MDB_DOMAIN)), "/chat/lark/sso"), []string{}
			if len(arg) == 0 {
				m.Cmd("web.chat./river").Table(func(index int, val map[string]string, head []string) {
					m.Cmd("web.chat./river", val[kit.MDB_HASH], chat.TOOL).Table(func(index int, value map[string]string, head []string) {
						list = append(list, kit.Keys(val[kit.MDB_NAME], value[kit.MDB_NAME]),
							cli.CMD, kit.Format([]string{HOME, val[kit.MDB_HASH], value[kit.MDB_HASH], val[kit.MDB_NAME] + "." + value[kit.MDB_NAME]}))
					})
				})
			} else {
				m.Option(ice.MSG_RIVER, arg[0])
				m.Option(ice.MSG_STORM, arg[1])
				link = kit.MergeURL(link, chat.RIVER, arg[0], chat.STORM, arg[1])
				m.Cmd("web.chat./river", arg[0], chat.TOOL, arg[1]).Table(func(index int, value map[string]string, head []string) {
					list = append(list, value[cli.CMD], cli.CMD, kit.Keys(value[cli.CTX], value[cli.CMD]))
				})
			}
			m.Cmd(FORM, CHAT_ID, m.Option(OPEN_CHAT_ID), name, text, "打开网页", "url", link, list)
		}},
	}})
}
