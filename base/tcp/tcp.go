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
				m.Cmd(HOST, aaa.WHITE, value[IP])
			})
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(CLIENT, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
				kit.Value(value, kit.Keym(kit.MDB_STATUS), CLOSE)
			})
			m.Richs(SERVER, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
				kit.Value(value, kit.Keym(kit.MDB_STATUS), CLOSE)
			})
			m.Save(PORT)
		}},
	},
}

func init() { ice.Index.Register(Index, nil, HOST, PORT, CLIENT, SERVER) }
