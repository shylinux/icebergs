package ctx

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _context_list(m *ice.Message, all bool) {
	m.Travel(func(p *ice.Context, s *ice.Context) {
		m.Push(kit.MDB_NAME, kit.Select("ice", s.Cap(ice.CTX_FOLLOW)))
		m.Push(ice.CTX_STATUS, s.Cap(ice.CTX_STATUS))
		m.Push(ice.CTX_STREAM, s.Cap(ice.CTX_STREAM))
		m.Push(kit.MDB_HELP, s.Help)
	})
}

const CONTEXT = "context"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			CONTEXT: {Name: "context name=web.chat action=context,command,config key auto", Help: "模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Search(kit.Select("ice", arg, 0)+".", func(p *ice.Context, s *ice.Context, key string) {
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
		},
	}, nil)
}
