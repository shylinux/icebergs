package ctx

import (
	"encoding/json"
	"os"
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _config_list(m *ice.Message) {
	for k, v := range m.Source().Configs {
		if k[0] == '/' || k[0] == '_' {
			continue // 内部配置
		}

		m.Push(kit.MDB_KEY, k)
		m.Push(kit.MDB_NAME, v.Name)
		m.Push(kit.MDB_VALUE, kit.Format(v.Value))
	}
	m.Sort(kit.MDB_KEY)
}
func _config_save(m *ice.Message, name string, arg ...string) {
	name = path.Join(m.Conf(CONFIG, kit.META_PATH), name)
	if f, p, e := kit.Create(name); m.Assert(e) {
		defer f.Close()

		msg := m.Spawn(m.Source())
		data := map[string]interface{}{}
		for _, k := range arg {
			data[k] = msg.Confv(k)
		}

		// 保存配置
		if s, e := json.MarshalIndent(data, "", "  "); m.Assert(e) {
			if n, e := f.Write(s); m.Assert(e) {
				m.Log_EXPORT(CONFIG, name, kit.MDB_FILE, p, kit.MDB_SIZE, n)
			}
		}
		m.Echo(p)
	}
}
func _config_load(m *ice.Message, name string, arg ...string) {
	name = path.Join(m.Conf(CONFIG, kit.META_PATH), name)
	if f, e := os.Open(name); e == nil {
		defer f.Close()

		msg := m.Spawn(m.Source())
		data := map[string]interface{}{}
		json.NewDecoder(f).Decode(&data)

		// 加载配置
		for k, v := range data {
			msg.Search(k, func(p *ice.Context, s *ice.Context, key string) {
				m.Log_IMPORT(CONFIG, kit.Keys(s.Name, key), kit.MDB_FILE, name)
				s.Configs[key].Value = v
			})
		}
	}
}
func _config_make(m *ice.Message, key string, arg ...string) {
	msg := m.Spawn(m.Source())
	if len(arg) > 1 {
		if strings.HasPrefix(arg[1], "@") {
			arg[1] = msg.Cmdx("nfs.cat", arg[1][1:])
		}
		// 修改配置
		msg.Confv(key, arg[0], kit.Parse(nil, "", arg[1:]...))
	}

	if len(arg) > 0 {
		m.Echo(kit.Formats(msg.Confv(key, arg[0])))
	} else {
		m.Echo(kit.Formats(msg.Confv(key)))
	}
}
func _config_rich(m *ice.Message, key string, sub string, arg ...string) {
	m.Rich(key, sub, kit.Data(arg))
}
func _config_grow(m *ice.Message, key string, sub string, arg ...string) {
	m.Grow(key, sub, kit.Dict(arg))
}

const (
	SAVE = "save"
	LOAD = "load"
	RICH = "rich"
	GROW = "grow"
)
const CONFIG = "config"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			CONFIG: {Name: CONFIG, Help: "配置", Value: kit.Data(kit.MDB_PATH, ice.VAR_CONF)},
		},
		Commands: map[string]*ice.Command{
			CONFIG: {Name: "config key auto", Help: "配置", Action: map[string]*ice.Action{
				SAVE: {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					_config_save(m, arg[0], arg[1:]...)
				}},
				LOAD: {Name: "load", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
					_config_load(m, arg[0], arg[1:]...)
				}},
				RICH: {Name: "rich", Help: "富有", Hand: func(m *ice.Message, arg ...string) {
					_config_rich(m, arg[0], arg[1], arg[2:]...)
				}},
				GROW: {Name: "grow", Help: "成长", Hand: func(m *ice.Message, arg ...string) {
					_config_grow(m, arg[0], arg[1], arg[2:]...)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					_config_list(m)
					return
				}
				_config_make(m, arg[0], arg[1:]...)
			}},
		},
	})
}
