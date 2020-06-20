package ctx

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"encoding/json"
	"os"
	"path"
	"strings"
)

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
	name = path.Join(msg.Conf(CONFIG, "meta.path"), name)
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
	name = path.Join(msg.Conf(CONFIG, "meta.path"), name)
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
func _config_rich(m *ice.Message, name string, key string, arg ...string) {
	m.Rich(name, key, kit.Dict(arg))
}
func _config_grow(m *ice.Message, name string, key string, arg ...string) {
	m.Grow(name, key, kit.Dict(arg))
}

const (
	CONTEXT = "context"
	COMMAND = "command"
	CONFIG  = "config"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CONFIG: {Name: "config", Help: "配置", Value: kit.Data("path", "var/conf")},
		},
		Commands: map[string]*ice.Command{
			CONFIG: {Name: "config [all] [chain [key [arg...]]]", Help: "配置", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if all, arg := _parse_arg_all(m, arg...); len(arg) == 0 {
					_config_list(m, all)
					return
				}

				switch arg[0] {
				case "save":
					_config_save(m, arg[1], arg[2:]...)
				case "load":
					_config_load(m, arg[1], arg[2:]...)
				case "grow":
					_config_grow(m, arg[1], arg[2], arg[3:]...)
				case "rich":
					_config_rich(m, arg[1], arg[2], arg[3:]...)
				default:
					_config_make(m, arg[0], arg[1:]...)
				}
			}},
		},
	}, nil)
}
