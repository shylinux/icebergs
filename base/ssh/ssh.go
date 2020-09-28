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
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		if _, ok := m.Target().Server().(*Frame); ok {
			m.Done()
		}
		m.Richs(CONNECT, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
			if value[kit.MDB_META] != nil {
				value = value[kit.MDB_META].(map[string]interface{})
			}
			kit.Value(value, "status", tcp.CLOSE)
		})
		m.Richs(SESSION, "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
			if value[kit.MDB_META] != nil {
				value = value[kit.MDB_META].(map[string]interface{})
			}
			kit.Value(value, "status", tcp.CLOSE)
		})
		m.Save()
	}},
}}

func init() {
	ice.Index.Register(Index, &Frame{},
		SOURCE, TARGET, PROMPT, RETURN,
		CONNECT, SESSION, SERVICE,
	)
}
