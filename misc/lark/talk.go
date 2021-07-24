package lark

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"
)

const TALK = "talk"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TALK: {Name: "talk text", Help: "聊天", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			cmds := kit.Split(strings.Join(arg, " "))
			if aaa.UserLogin(m, m.Option(OPEN_ID), ""); !m.Right(cmds) {
				if aaa.UserLogin(m, m.Option(OPEN_CHAT_ID), ""); !m.Right(cmds) {
					m.Cmd(DUTY, m.Option(APP_ID), m.Option(OPEN_CHAT_ID), m.Option("text_without_at_bot"))
					m.Cmd(HOME)
					return // 没有权限
				}
			}

			// 执行命令
			m.Cmdy(cmds)
		}},
	}})
}
