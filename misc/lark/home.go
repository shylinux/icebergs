package lark

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

const HOME = "home"

func init() {
	Index.MergeCommands(ice.Commands{
		HOME: {Name: "home river storm title content", Help: "首页", Hand: func(m *ice.Message, arg ...string) {
			name := kit.Select(m.Option(ice.MSG_USERNAME), m.Option(ice.MSG_USERNICK))
			kit.If(len(name) > 10, func() { name = name[:10] })
			name += "的" + kit.Select("应用列表", arg, 2)
			text, link, list := kit.Select("", arg, 3), kit.MergeURL2(mdb.Conf(m, web.SHARE, kit.Keym(web.DOMAIN)), "/chat/lark/sso"), []string{}
			if len(arg) == 0 {
				m.Cmd("web.chat./river", func(val ice.Maps) {
					m.Cmd("web.chat./river", val[mdb.HASH], chat.STORM, func(value ice.Maps) {
						list = append(list, kit.Keys(val[mdb.NAME], value[mdb.NAME]),
							ice.CMD, kit.Format([]string{HOME, val[mdb.HASH], value[mdb.HASH], val[mdb.NAME] + nfs.PT + value[mdb.NAME]}))
					})
				})
			} else {
				m.Options(ice.MSG_RIVER, arg[0], ice.MSG_STORM, arg[1])
				link = kit.MergeURL(link, chat.RIVER, arg[0], chat.STORM, arg[1])
				m.Cmd("web.chat./river", arg[0], chat.STORM, arg[1], func(value ice.Maps) {
					list = append(list, value[ice.CMD], ice.CMD, kit.Keys(value[ice.CTX], value[ice.CMD]))
				})
			}
			m.Cmd(FORM, CHAT_ID, m.Option(OPEN_CHAT_ID), name, text, "打开网页", "url", link, list)
		}},
	})
}
