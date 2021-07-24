package lark

import (
	"math/rand"
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const RAND = "rand"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		RAND: {Name: "rand", Help: "随机", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			msg := m.Cmd(GROUP, m.Option(APP_ID), EMPLOYEE, m.Option(OPEN_CHAT_ID))
			list := msg.Appendv(kit.MDB_NAME)
			if strings.Contains(m.Option(CONTENT), "誰") {
				m.Echo(strings.Replace(m.Option(CONTENT), "誰", list[rand.Intn(len(list))], 1))
				return
			}
			m.Echo(list[rand.Intn(len(list))])
		}},
	}})
}
