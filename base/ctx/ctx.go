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

func _parse_arg_all(m *ice.Message, arg ...string) (bool, []string) {
	if len(arg) > 0 && arg[0] == "all" {
		return true, arg[1:]
	}
	return false, arg

}
func _parse_arg_chain(m *ice.Message, arg ...string) (string, []string) {
	if len(arg) > 1 {
		return kit.Keys(arg[0], arg[1]), arg[2:]
	}
	return arg[0], arg[1:]
}

func _config_list(m *ice.Message, all bool) {
	p := m.Spawn(m.Source())
	if all {
		p = ice.Pulse
	}
	p.Travel(func(p *ice.Context, s *ice.Context, key string, conf *ice.Config) {
		m.Push("key", key)
		m.Push("name", conf.Name)
		m.Push("value", kit.Format(conf.Value))
	})

}
func _config_save(m *ice.Message, name string, arg ...string) {
	msg := m.Spawn(m.Source())
	// 保存配置
	if m.Cap(ice.CTX_STATUS) != ice.ICE_START {
		return
	}
	name = path.Join(msg.Conf(ice.CTX_CONFIG, "meta.path"), name)
	if f, p, e := kit.Create(name); m.Assert(e) {
		data := map[string]interface{}{}
		for _, k := range arg {
			data[k] = msg.Confv(k)
		}
		if s, e := json.MarshalIndent(data, "", "  "); m.Assert(e) {
			if n, e := f.Write(s); m.Assert(e) {
				m.Log("info", "save %d %s", n, p)
			}
		}
		m.Echo(p)
	}
}
func _config_load(m *ice.Message, name string, arg ...string) {
	msg := m.Spawn(m.Source())
	// 加载配置
	name = path.Join(msg.Conf(ice.CTX_CONFIG, "meta.path"), name)
	if f, e := os.Open(name); e == nil {
		data := map[string]interface{}{}
		json.NewDecoder(f).Decode(&data)

		for k, v := range data {
			msg.Search(k, func(p *ice.Context, s *ice.Context, key string) {
				m.Log("info", "load %s.%s %v", s.Name, key, kit.Format(v))
				s.Configs[key].Value = v
			})
		}
	}
}
func _config_make(m *ice.Message, chain string, arg ...string) {
	msg := m.Spawn(m.Source())
	if len(arg) > 1 {
		if strings.HasPrefix(arg[1], "@") {
			msg.Conf(chain, arg[0], msg.Cmdx("nfs.cat", arg[1][1:]))
		} else {
			msg.Conf(chain, arg[0], kit.Parse(nil, "", arg[1:]...))
		}
	}

	if len(arg) > 0 {
		// 读取配置
		m.Echo(kit.Formats(msg.Confv(chain, arg[0])))
	} else {
		// 读取配置
		m.Echo(kit.Formats(msg.Confv(chain)))
	}
}

func _command_list(m *ice.Message, all bool) {
	p := m.Spawn(m.Source())
	if all {
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
			m.Push("name", kit.Format(v.Name))
			m.Push("help", kit.Simple(v.Help)[0])
			// m.Push("list", kit.Format(v.List))
		}
	})
}
func _command_make(m *ice.Message, cmd *ice.Command) {
	var list []string
	switch name := cmd.Name.(type) {
	case []string, []interface{}:
		list = kit.Split(kit.Simple(name)[0])
	default:
		list = kit.Split(strings.Split(kit.Format(name), ";")[0])
	}

	button := false
	for i, v := range list {
		if i > 0 {
			switch ls := kit.Split(v, ":="); ls[0] {
			case "[", "]":
			case "auto":
				cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "button", "name", "查看", "value", "auto")...)
				cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, "button", "name", "返回", "value", "Last")...)
				button = true
			default:
				kind, value := "text", ""
				if len(ls) == 3 {
					kind, value = ls[1], ls[2]
				} else if len(ls) == 2 {
					if strings.Contains(v, "=") {
						value = ls[1]
					} else {
						kind = ls[1]
					}
				}
				if kind == "button" {
					button = true
				}
				cmd.List = append(cmd.List, kit.List(kit.MDB_INPUT, kind, "name", ls[0], "value", value)...)
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
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		ice.CTX_CONFIG: {Name: "config", Help: "配置", Value: kit.Data("path", "var/conf")},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_CONTEXT: {Name: "context [all]", Help: "模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if all, arg := _parse_arg_all(m, arg...); len(arg) == 0 {
				_context_list(m, all)
				return
			}

			if len(arg) == 1 {
				m.Cmdy(ice.CTX_COMMAND, arg[0]+".")
			} else {
				m.Search(arg[0]+".", func(p *ice.Context, s *ice.Context, key string) {
					msg := m.Spawn(s)
					switch arg[1] {
					case "command":
						msg.Cmdy(ice.CTX_COMMAND, arg[0], arg[2:])
					case "config":
						msg.Cmdy(ice.CTX_CONFIG, arg[2:])
					}
					m.Copy(msg)
				})
			}

		}},
		ice.CTX_COMMAND: {Name: "command [all] [context [command run arg...]]", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if all, arg := _parse_arg_all(m, arg...); len(arg) == 0 {
				_command_list(m, all)
				return
			}

			chain, arg := _parse_arg_chain(m, arg...)
			m.Search(chain, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
				if len(arg) == 0 {
					// 命令列表
					m.Push("key", key)
					m.Push("name", cmd.Name)
					m.Push("help", kit.Simple(cmd.Help)[0])
					m.Push("meta", kit.Format(cmd.Meta))
					if len(cmd.List) == 0 {
						_command_make(m, cmd)
					}
					m.Push("list", kit.Format(cmd.List))
				} else {
					if you := m.Option(kit.Format(kit.Value(cmd.Meta, "remote"))); you != "" {
						// 远程命令
						m.Copy(m.Spawns(s).Cmd(ice.WEB_SPACE, you, ice.CTX_COMMAND, chain, "run", arg[1:]))
					} else {
						// 本地命令
						m.Copy(s.Run(m.Spawns(s), cmd, key, arg[1:]...))
					}
				}
			})
		}},
		ice.CTX_CONFIG: {Name: "config [all] [chain [key [arg...]]]", Help: "配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if all, arg := _parse_arg_all(m, arg...); len(arg) == 0 {
				_config_list(m, all)
				return
			}

			switch arg[0] {
			case "save":
				_config_save(m, arg[1], arg[2:]...)
			case "load":
				_config_load(m, arg[1], arg[2:]...)
			default:
				_config_make(m, arg[0], arg[1:]...)
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
