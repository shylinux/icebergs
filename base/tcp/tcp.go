package tcp

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"
)

const TCP = "tcp"

var Index = &ice.Context{Name: TCP, Help: "通信模块",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			m.Cmd(HOST).Table(func(index int, value map[string]string, head []string) {
				m.Cmd(HOST, aaa.White, value["ip"])
			})
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(CLIENT, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
				kit.Value(value, "meta.status", CLOSE)
			})
			m.Richs(SERVER, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
				kit.Value(value, "meta.status", CLOSE)
			})
			m.Save()
		}},
	},
}

func init() { ice.Index.Register(Index, nil, HOST, PORT, CLIENT, SERVER) }
