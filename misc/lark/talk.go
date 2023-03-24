package lark

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const TALK = "talk"

func init() {
	Index.MergeCommands(ice.Commands{
		TALK: {Name: "talk text", Help: "聊天", Hand: func(m *ice.Message, arg ...string) {
			cmds := kit.Split(strings.Join(arg, " "))
			if aaa.SessAuth(m, kit.Dict(aaa.USERNAME, m.Option(OPEN_ID))); !aaa.Right(m, cmds) {
				if aaa.SessAuth(m, kit.Dict(aaa.USERNAME, m.Option(OPEN_CHAT_ID))); !aaa.Right(m, cmds) {
					m.Cmd(DUTY, m.Option(OPEN_CHAT_ID), m.Option("text_without_at_bot"))
					m.Cmd(HOME)
					return
				}
			}
			if m.Cmdy(cmds); m.Result() != "" && !m.IsErrNotFound() {
				m.Cmd(SEND, m.Option(APP_ID), m.Option(OPEN_CHAT_ID), m.Result())
				return
			} else if m.Length() == 0 {
				m.Set(ice.MSG_RESULT)
				m.Cmdy(cli.SYSTEM, cmds)
				m.Cmd(SEND, m.Option(APP_ID), m.Option(OPEN_CHAT_ID), m.Result())
				return
			}
			val := []string{}
			m.Table(func(index int, value ice.Maps, head []string) {
				kit.For(head, func(k string) { val = append(val, kit.Format("%s:\t%s", k, value[k])) })
				val = append(val, ice.NL)
			})
			_lark_post(m, m.Option(APP_ID), "/open-apis/message/v4/send/", web.SPIDE_DATA, kit.Formats(
				kit.Dict("msg_type", "interactive", "chat_id", m.Option(OPEN_CHAT_ID), "card", kit.Dict(
					"header", kit.Dict("title", kit.Dict("tag", "lark_md", "content", strings.Join(cmds, " "))),
					"elements", []ice.Any{kit.Dict("tag", "div", "fields", []ice.Any{
						kit.Dict("is_short", true, "text", kit.Dict("tag", "lark_md", "content", strings.Join(val, ice.NL))),
					})},
				)),
			))
		}},
	})
}
