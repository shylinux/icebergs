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
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TALK: {Name: "talk text", Help: "聊天", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			cmds := kit.Split(strings.Join(arg, " "))
			if aaa.UserLogin(m, m.Option(OPEN_ID), ""); !m.Right(cmds) {
				if aaa.UserLogin(m, m.Option(OPEN_CHAT_ID), ""); !m.Right(cmds) {
					m.Cmd(DUTY, m.Option(OPEN_CHAT_ID), m.Option("text_without_at_bot"))
					m.Cmd(HOME)
					return // 没有权限
				}
			}

			// 执行命令
			if m.Cmdy(cmds); m.Result() != "" && m.Result(1) != ice.ErrNotFound {
				m.Cmd(SEND, m.Option(APP_ID), m.Option(OPEN_CHAT_ID), m.Result())
				return
			}
			if m.Length() == 0 {
				m.Set(ice.MSG_RESULT)
				m.Cmdy(cli.SYSTEM, cmds)
				m.Cmd(SEND, m.Option(APP_ID), m.Option(OPEN_CHAT_ID), m.Result())
				return
			}

			val := []string{}
			m.Table(func(index int, value map[string]string, head []string) {
				for _, key := range head {
					val = append(val, kit.Format("%s:\t%s", key, value[key]))
				}
				val = append(val, "\n")
			})

			_lark_post(m, m.Option(APP_ID), "/open-apis/message/v4/send/", web.SPIDE_DATA, kit.Formats(
				kit.Dict("msg_type", "interactive", "chat_id", m.Option(OPEN_CHAT_ID), "card", kit.Dict(
					"header", kit.Dict("title", kit.Dict("tag", "lark_md", "content", strings.Join(cmds, " "))),
					"elements", []interface{}{kit.Dict("tag", "div", "fields", []interface{}{
						kit.Dict("is_short", true, "text", kit.Dict(
							"tag", "lark_md", "content", strings.Join(val, "\n"),
						)),
					})},
				)),
			))
		}},
	}})
}
