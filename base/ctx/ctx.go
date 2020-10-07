package ctx

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"strings"
)

var Index = &ice.Context{Name: "ctx", Help: "配置模块",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(mdb.SEARCH, mdb.CREATE, COMMAND, m.AddCmd(&ice.Command{Hand: func(m *ice.Message, c *ice.Context, cc string, arg ...string) {
				arg = arg[1:]
				ice.Pulse.Travel(func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
					if strings.HasPrefix(key, "_") || strings.HasPrefix(key, "/") {
						return
					}
					if arg[1] != "" && arg[1] != key && arg[1] != s.Name {
						return
					}
					if arg[2] != "" && !strings.Contains(kit.Format(cmd.Name), arg[2]) && !strings.Contains(kit.Format(cmd.Help), arg[2]) {
						return
					}

					m.Push("pod", "")
					m.Push("ctx", "web.chat")
					m.Push("cmd", cc)

					m.Push("time", m.Time())
					m.Push("size", "")

					m.Push("type", COMMAND)
					m.Push("name", key)
					m.Push("text", s.Cap(ice.CTX_FOLLOW))
				})
			}}))
		}},
	},
}

func init() { ice.Index.Register(Index, nil, CONTEXT, COMMAND, CONFIG) }
