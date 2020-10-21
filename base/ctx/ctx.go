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
				fields := kit.Split(m.Option(mdb.FIELDS))
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

					for _, k := range fields {
						switch k {
						case kit.SSH_POD:
							m.Push(k, m.Option(ice.MSG_USERPOD))
						case kit.SSH_CTX:
							m.Push(k, m.Prefix())
						case kit.SSH_CMD:
							m.Push(k, "_")
						case kit.MDB_TIME:
							m.Push(k, m.Time())
						case kit.MDB_SIZE:
							m.Push(k, len(cmd.List))
						case kit.MDB_TYPE:
							m.Push(k, COMMAND)
						case kit.MDB_NAME:
							m.Push(k, key)
						case kit.MDB_TEXT:
							m.Push(k, m.Prefix())
						default:
							m.Push(k, "")
						}
					}
				})
			}}))
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
	},
}

func init() { ice.Index.Register(Index, nil, CONTEXT, COMMAND, CONFIG) }
