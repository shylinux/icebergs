package ctx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"encoding/json"
	"os"
	"path"
	"sort"
)

var Index = &ice.Context{Name: "ctx", Help: "元始模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.CTX_CONFIG: {Name: "config", Help: "配置", Value: kit.Data("path", "var/conf")},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_CONTEXT: {Name: "context", Help: "模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			ice.Pulse.Travel(func(p *ice.Context, s *ice.Context) {
				if p != nil {
					m.Push("ups", p.Name)
				} else {
					m.Push("ups", "shy")
				}
				m.Push("name", s.Name)
				m.Push(ice.CTX_STATUS, s.Cap(ice.CTX_STATUS))
				m.Push(ice.CTX_STREAM, s.Cap(ice.CTX_STREAM))
				m.Push("help", s.Help)
			})
		}},
		ice.CTX_COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				ice.Pulse.Travel(func(p *ice.Context, s *ice.Context) {
					list := []string{}
					for k := range s.Commands {
						if k[0] == '/' || k[0] == '_' {
							continue
						}
						list = append(list, k)
					}
					sort.Strings(list)

					for _, k := range list {
						v := s.Commands[k]
						if p != nil && p != ice.Index {
							m.Push("key", p.Name+"."+s.Name)
						} else {
							m.Push("key", s.Name)
						}
						m.Push("index", k)
						m.Push("name", v.Name)
						m.Push("help", v.Help)
					}
				})
				return
			}

			chain := arg[0]
			if len(arg) > 1 {
				chain = arg[0] + "." + arg[1]
				arg = arg[1:]
			}
			arg = arg[1:]
			m.Search(chain, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
				if len(arg) == 0 {
					m.Push("name", cmd.Name)
					m.Push("help", cmd.Help)
					m.Push("meta", kit.Format(cmd.Meta))
					m.Push("list", kit.Format(cmd.List))
				} else {
					if you := m.Option(kit.Format(kit.Value(cmd.Meta, "remote"))); you != "" {
						m.Copy(m.Spawns(s).Cmd("web.space", you, "ctx.command", chain, "run", arg[1:]))
					} else {
						m.Copy(s.Run(m.Spawns(s), cmd, key, arg[1:]...))
					}
				}
			})
		}},
		ice.CTX_CONFIG: {Name: "config", Help: "配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				ice.Pulse.Travel(func(p *ice.Context, s *ice.Context, key string, conf *ice.Config) {
					m.Push("key", key)
					m.Push("name", conf.Name)
					m.Push("value", kit.Format(conf.Value))
				})
				return
			}

			switch arg[0] {
			case "save":
				arg[1] = path.Join(m.Conf(ice.CTX_CONFIG, ice.Meta("path")), arg[1])
				if f, p, e := kit.Create(arg[1]); m.Assert(e) {
					data := map[string]interface{}{}
					for _, k := range arg[2:] {
						data[k] = m.Confv(k)
					}
					if s, e := json.MarshalIndent(data, "", "  "); m.Assert(e) {
						if n, e := f.Write(s); m.Assert(e) {
							m.Log("info", "save %d %s", n, p)
						}
					}
					m.Echo(p)
				}
			case "load":
				arg[1] = path.Join(m.Conf(ice.CTX_CONFIG, ice.Meta("path")), arg[1])
				if f, e := os.Open(arg[1]); e == nil {
					data := map[string]interface{}{}
					json.NewDecoder(f).Decode(&data)

					for k, v := range data {
						m.Search(k, func(p *ice.Context, s *ice.Context, key string) {
							m.Log("info", "load %s.%s %v", s.Name, key, kit.Format(v))
							s.Configs[key].Value = v
						})
					}
				}
			default:
				m.Echo(kit.Formats(m.Confv(arg[0])))
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
