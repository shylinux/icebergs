package ctx

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _context_list(m *ice.Message, all bool) {
	m.Travel(func(p *ice.Context, s *ice.Context) {
		m.Push(kit.MDB_NAME, s.Cap(ice.CTX_FOLLOW))
		m.Push(kit.MDB_STATUS, s.Cap(ice.CTX_STATUS))
		m.Push(kit.MDB_STREAM, s.Cap(ice.CTX_STREAM))
		m.Push(kit.MDB_HELP, s.Help)
	})
}

const CONTEXT = "context"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		CONTEXT: {Name: "context name=web.chat action=context,command,config key auto", Help: "模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Search(kit.Select(ice.ICE, arg, 0)+ice.PT, func(p *ice.Context, s *ice.Context, key string) {
				msg := m.Spawn(s)
				defer m.Copy(msg)

				switch kit.Select(CONTEXT, arg, 1) {
				case CONTEXT:
					_context_list(msg, true)
				case COMMAND:
					msg.Cmdy(COMMAND, arg[2:])
				case CONFIG:
					msg.Cmdy(CONFIG, arg[2:])
				}
			})
		}},
	}})
}
