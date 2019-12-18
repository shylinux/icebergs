package ctx

import (
	"encoding/json"
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
	"os"
	"sort"
)

var Index = &ice.Context{Name: "ctx", Help: "元始模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		"context": {Name: "context", Help: "模块", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			ice.Pulse.Travel(func(p *ice.Context, s *ice.Context) {
				if p != nil {
					m.Push("ups", p.Name)
				} else {
					m.Push("ups", "shy")
				}
				m.Push("name", s.Name)
				m.Push("status", s.Cap("status"))
				m.Push("stream", s.Cap("stream"))
				m.Push("help", s.Help)
			})
		}},
		"command": {Name: "command", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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

			m.Search(arg[0]+"."+arg[1], func(p *ice.Context, s *ice.Context, key string) {
				if i, ok := s.Commands[key]; ok {
					if len(arg) == 2 {
						m.Push("name", i.Name)
						m.Push("help", i.Help)
						m.Push("meta", kit.Format(i.Meta))
						m.Push("list", kit.Format(i.List))
					} else {
						if i.Meta != nil && kit.Format(i.Meta["remote"]) == "true" && m.Option("you") != "" {
							m.Copy(m.Spawns(s).Cmd("web.space", m.Option("you"), "ctx.command", arg[0], arg[1], "run", arg[3:]))
						} else {
							m.Copy(m.Spawns(s).Runs(key, key, arg[3:]...))
						}
					}
				}
			})
		}},
		"config": {Name: "config", Help: "配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch arg[0] {
			case "save":
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
			}
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
