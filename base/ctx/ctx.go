package ctx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

func _parse_arg_all(m *ice.Message, arg ...string) (bool, []string) {
	if len(arg) > 0 && arg[0] == "all" {
		return true, arg[1:]
	}
	return false, arg
}

func _context_list(m *ice.Message, all bool) {
	p := m.Spawn(m.Source())
	if all {
		p = ice.Pulse
	}

	p.Travel(func(p *ice.Context, s *ice.Context) {
		if p != nil {
			m.Push("ups", kit.Select("shy", p.Cap(ice.CTX_FOLLOW)))
		} else {
			m.Push("ups", "shy")
		}
		m.Push("name", s.Name)
		m.Push(ice.CTX_STATUS, s.Cap(ice.CTX_STATUS))
		m.Push(ice.CTX_STREAM, s.Cap(ice.CTX_STREAM))
		m.Push("help", s.Help)
	})
}

var Index = &ice.Context{Name: "ctx", Help: "配置模块",
	Commands: map[string]*ice.Command{
		CONTEXT: {Name: "context [all] [name [command|config arg...]]", Help: "模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if all, arg := _parse_arg_all(m, arg...); len(arg) == 0 {
				_context_list(m, all)
				return
			}

			if len(arg) == 1 {
				m.Cmdy(COMMAND, arg[0])
			} else {
				m.Search(arg[0]+".", func(p *ice.Context, s *ice.Context, key string) {
					msg := m.Spawn(s)
					switch arg[1] {
					case "command":
						msg.Cmdy(COMMAND, arg[0], arg[2:])
					case "config":
						msg.Cmdy(CONFIG, arg[2:])
					}
					m.Copy(msg)
				})
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil, CONTEXT, COMMAND, CONFIG) }
