package ssh

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"
)

const SSH = "ssh"

var Index = &ice.Context{Name: SSH, Help: "终端模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Load()
		m.Richs(SESSION, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
			kit.Value(value, kit.Keym(kit.MDB_STATUS), tcp.CLOSE)
		})
		m.Richs(CHANNEL, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
			kit.Value(value, kit.Keym(kit.MDB_STATUS), tcp.CLOSE)
		})
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if f, ok := m.Target().Server().(*Frame); ok {
			f.close()
		}
		m.Save()
	}},
}}

func init() {
	ice.Index.Register(Index, &Frame{},
		CONNECT, SESSION, SERVICE, CHANNEL,
		SOURCE, TARGET, PROMPT, PRINTF, SCREEN, RETURN,
	)
}
