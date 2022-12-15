package lark

import (
	"math/rand"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

const RAND = "rand"

func init() {
	Index.MergeCommands(ice.Commands{
		RAND: {Name: "rand", Help: "随机", Hand: func(m *ice.Message, arg ...string) {
			msg := m.Cmd(GROUP, m.Option(APP_ID), EMPLOYEE, m.Option(OPEN_CHAT_ID))
			if list := msg.Appendv(mdb.NAME); strings.Contains(m.Option(CONTENT), "誰") {
				m.Echo(strings.Replace(m.Option(CONTENT), "誰", list[rand.Intn(len(list))], 1))
			} else {
				m.Echo(list[rand.Intn(len(list))])
			}
		}},
	})
}
