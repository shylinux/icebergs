package ctx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"encoding/json"
	"os"
	"path"
	"sort"
	"strings"
)

var Index = &ice.Context{Name: "ctx", Help: "配置模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.CTX_CONFIG: {Name: "config", Help: "配置", Value: kit.Data("path", "var/conf")},
		"demo":         {Name: "demo", Help: "配置", Value: kit.Data("path", "var/conf")},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_CONTEXT: {Name: "context [all]", Help: "模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			if p := m.Spawn(m.Source()); len(arg) == 0 {
				if all == true {
					p = ice.Pulse
				}
				// 模块列表
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
				return
			} else if len(arg) == 1 {
				m.Cmdy(ice.CTX_COMMAND, arg[0]+".")
			} else {
				m.Search(arg[0]+".", func(p *ice.Context, s *ice.Context, key string) {
					msg := m.Spawn(s)
					switch arg[1] {
					case "command":
						msg.Cmdy(ice.CTX_COMMAND, arg[0], arg[2:])
					case "config":
						msg.Cmdy(ice.CTX_CONFIG, arg[2:])
					case "cache":
					}
					m.Copy(msg)
				})
			}

		}},
		ice.CTX_COMMAND: {Name: "command [all] [context [command run arg...]]", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			if p := m.Spawn(m.Source()); len(arg) == 0 {
				if all == true {
					p = ice.Pulse
				}
				// 命令列表
				p.Travel(func(p *ice.Context, s *ice.Context) {
					list := []string{}
					for k := range s.Commands {
						if k[0] == '/' || k[0] == '_' {
							// 内部命令
							continue
						}
						list = append(list, k)
					}
					sort.Strings(list)

					for _, k := range list {
						v := s.Commands[k]
						m.Push("key", s.Cap(ice.CTX_FOLLOW))
						m.Push("index", k)
						m.Push("name", v.Name)
						m.Push("help", kit.Simple(v.Help)[0])
						m.Push("list", kit.Format(v.List))
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
					// 命令列表
					m.Push("key", key)
					m.Push("name", cmd.Name)
					m.Push("help", kit.Simple(cmd.Help)[0])
					m.Push("meta", kit.Format(cmd.Meta))
					if len(cmd.List) == 0 {
						list := kit.Split(cmd.Name)
						button := false
						for i, v := range list {
							if i > 0 {
								ls := kit.Split(v, ":=")
								switch ls[0] {
								case "auto":
									cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "button", "name", "查看", "value", "auto")...)
									cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "button", "name", "返回", "value", "Last")...)
									button = true
								default:
									if len(ls) > 1 && ls[1] == "button" {
										button = true
									}
									cmd.List = append(cmd.List, kit.List(
										kit.MDB_INPUT, kit.Select("text", ls, 1), "name", ls[0], "value", kit.Select("", ls, 2),
									)...)
								}
							}
						}
						if len(cmd.List) == 0 {
							cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "text", "name", "name")...)
						}
						if !button {
							cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "button", "name", "查看")...)
							cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "button", "name", "返回", "value", "Last")...)
						}
					}
					m.Push("list", kit.Format(cmd.List))
				} else {
					if you := m.Option(kit.Format(kit.Value(cmd.Meta, "remote"))); you != "" {
						// 远程命令
						m.Copy(m.Spawns(s).Cmd(ice.WEB_SPACE, you, "ctx.command", chain, "run", arg[1:]))
					} else {
						// 本地命令
						m.Copy(s.Run(m.Spawns(s), cmd, key, arg[1:]...))
					}
				}
			})
		}},
		ice.CTX_CONFIG: {Name: "config [all] [save|load] chain key arg...", Help: "配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			all := false
			if len(arg) > 0 && arg[0] == "all" {
				all, arg = true, arg[1:]
			}

			msg := m.Spawn(m.Source())
			if len(arg) == 0 {
				if all == true {
					msg = ice.Pulse
				}
				// 配置列表
				msg.Travel(func(p *ice.Context, s *ice.Context, key string, conf *ice.Config) {
					m.Push("key", key)
					m.Push("name", conf.Name)
					m.Push("value", kit.Format(conf.Value))
				})
				return
			}

			switch arg[0] {
			case "save":
				// 保存配置
				if m.Cap(ice.CTX_STATUS) != ice.ICE_START {
					break
				}
				arg[1] = path.Join(msg.Conf(ice.CTX_CONFIG, "meta.path"), arg[1])
				if f, p, e := kit.Create(arg[1]); m.Assert(e) {
					data := map[string]interface{}{}
					for _, k := range arg[2:] {
						data[k] = msg.Confv(k)
					}
					if s, e := json.MarshalIndent(data, "", "  "); m.Assert(e) {
						if n, e := f.Write(s); m.Assert(e) {
							m.Log("info", "save %d %s", n, p)
						}
					}
					m.Echo(p)
				}
			case "load":
				// 加载配置
				arg[1] = path.Join(msg.Conf(ice.CTX_CONFIG, "meta.path"), arg[1])
				if f, e := os.Open(arg[1]); e == nil {
					data := map[string]interface{}{}
					json.NewDecoder(f).Decode(&data)

					for k, v := range data {
						msg.Search(k, func(p *ice.Context, s *ice.Context, key string) {
							m.Log("info", "load %s.%s %v", s.Name, key, kit.Format(v))
							s.Configs[key].Value = v
						})
					}
				}
			default:
				if len(arg) > 2 {
					if strings.HasPrefix(arg[2], "@") {
						msg.Conf(arg[0], arg[1], msg.Cmdx("nfs.cat", arg[2][1:]))
					} else {
						msg.Conf(arg[0], arg[1], kit.Parse(nil, "", arg[2:]...))
					}

				}
				if len(arg) > 1 {
					// 读取配置
					m.Echo(kit.Formats(msg.Confv(arg[0], arg[1])))
				} else {
					// 读取配置
					m.Echo(kit.Formats(msg.Confv(arg[0])))
				}
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
