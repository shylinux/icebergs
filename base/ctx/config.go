package ctx

import (
	"encoding/json"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/miss"
)

func _config_save(m *ice.Message, name string, arg ...string) {
	name = path.Join(ice.VAR_CONF, name)
	if f, p, e := miss.CreateFile(name); m.Assert(e) {
		defer f.Close()

		msg := m.Spawn(m.Source())
		data := ice.Map{}
		for _, k := range arg {
			if v := msg.Confv(k); v != "" {
				data[k] = v
			}
		}

		// 保存配置
		if s, e := json.MarshalIndent(data, "", "  "); m.Assert(e) {
			if _, e := f.Write(s); m.Assert(e) {
			}
		}
		m.Echo(p)
	}
}
func _config_load(m *ice.Message, name string, arg ...string) {
	name = path.Join(ice.VAR_CONF, name)
	if f, e := miss.OpenFile(name); e == nil {
		defer f.Close()

		msg := m.Spawn(m.Source())
		data := ice.Map{}
		json.NewDecoder(f).Decode(&data)

		// 加载配置
		for k, v := range data {
			msg.Search(k, func(p *ice.Context, s *ice.Context, key string) {
				if s.Configs[key] == nil {
					s.Configs[key] = &ice.Config{}
				}
				s.Configs[key].Value = v
			})
		}
	}
}
func _config_make(m *ice.Message, key string, arg ...string) {
	msg := m.Spawn(m.Source())
	if len(arg) > 1 {
		if strings.HasPrefix(arg[1], "@") {
			arg[1] = msg.Cmdx(nfs.CAT, arg[1][1:])
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
func _config_list(m *ice.Message) {
	for k, v := range m.Source().Configs {
		if k[0] == '/' || k[0] == '_' {
			continue // 内部配置
		}

		m.Push(mdb.KEY, k)
		m.Push(mdb.NAME, v.Name)
		m.Push(mdb.VALUE, kit.Format(v.Value))
	}
	m.Sort(mdb.KEY)
}

const (
	SAVE = "save"
	LOAD = "load"
	RICH = "rich"
	GROW = "grow"
)
const CONFIG = "config"

func init() {
	Index.MergeCommands(ice.Commands{
		CONFIG: {Name: "config key auto reset", Help: "配置", Actions: ice.Actions{
			SAVE: {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				_config_save(m, arg[0], arg[1:]...)
			}},
			LOAD: {Name: "load", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
				_config_load(m, arg[0], arg[1:]...)
			}},
			RICH: {Name: "rich", Help: "富有", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, arg[0], arg[1], mdb.HASH, arg[2:])
			}},
			GROW: {Name: "grow", Help: "成长", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, arg[0], arg[1], mdb.LIST, arg[2:])
			}},
			"list": {Name: "list", Help: "列表", Hand: func(m *ice.Message, arg ...string) {
				list := []ice.Any{}
				for _, v := range arg[2:] {
					list = append(list, v)
				}
				m.Confv(arg[0], arg[1], kit.List(list...))
			}},
			mdb.REMOVE: {Name: "remove key sub", Hand: func(m *ice.Message, arg ...string) {
				m.Conf(m.Option("key"), m.Option("sub"), "")
				m.Go(func() { m.Cmd(ice.EXIT, 1) })
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_config_list(m)
				return
			}
			_config_make(m, arg[0], arg[1:]...)
			DisplayStoryJSON(m)
		}},
	})
}
func init() {
	ice.Info.Save = Save
	ice.Info.Load = Load
}
func Save(m *ice.Message, arg ...string) *ice.Message {
	if len(arg) == 0 {
		for k := range m.Target().Configs {
			arg = append(arg, k)
		}
	}
	for i, k := range arg {
		arg[i] = m.Prefix(k)
	}
	return m.Cmd(CONFIG, SAVE, m.Prefix(nfs.JSON), arg)
}
func Load(m *ice.Message, arg ...string) *ice.Message {
	if len(arg) == 0 {
		for k := range m.Target().Configs {
			arg = append(arg, k)
		}
	}
	for i, k := range arg {
		arg[i] = m.Prefix(k)
	}
	return m.Cmd(CONFIG, LOAD, m.Prefix(nfs.JSON), arg)
}
func ConfAction(args ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: mdb.AutoConfig(args...)}
}
func ConfigFromOption(m *ice.Message, arg ...string) {
	for _, k := range arg {
		m.Config(k, kit.Select(m.Config(k), m.Option(k)))
	}
}
func ConfigAuto(m *ice.Message, arg ...string) {
	if cs := m.Target().Configs; cs[m.CommandKey()] == nil {
		cs[m.CommandKey()] = &ice.Config{Value: kit.Data()}
		ice.Info.Load(m, m.CommandKey())
	}
}
