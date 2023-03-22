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

func _config_only(v ice.Any, arg ...string) bool {
	switch v := v.(type) {
	case nil:
		return true
	case ice.Map:
		if len(v) > len(arg) {
			return false
		}
		for k := range v {
			if kit.IndexOf(arg, k) == -1 {
				return false
			}
		}
		return true
	}
	return false
}
func _config_save(m *ice.Message, name string, arg ...string) {
	data, msg := ice.Map{}, m.Spawn(m.Source())
	for _, k := range arg {
		if v := mdb.Confv(msg, k); _config_only(v, mdb.META) {
			continue
		} else {
			data[k] = v
		}
	}
	if len(data) == 0 {
		return
	}
	if f, _, e := miss.CreateFile(path.Join(ice.VAR_CONF, name)); m.Assert(e) {
		defer f.Close()
		if s, e := json.MarshalIndent(data, "", "  "); !m.Warn(e) {
			if _, e := f.Write(s); !m.Warn(e) {
			}
		}
	}
}
func _config_load(m *ice.Message, name string, arg ...string) {
	if f, e := miss.OpenFile(path.Join(ice.VAR_CONF, name)); e == nil {
		defer f.Close()
		data, msg := ice.Map{}, m.Spawn(m.Source())
		json.NewDecoder(f).Decode(&data)
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
		if strings.HasPrefix(arg[1], ice.AT) {
			arg[1] = msg.Cmdx(nfs.CAT, arg[1][1:])
		}
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
		if IsOrderCmd(k) {
			continue
		}
		m.Push(mdb.KEY, k).Push(mdb.NAME, v.Name).Push(mdb.VALUE, kit.Format(v.Value))
	}
	m.Sort(mdb.KEY)
}

const (
	SAVE = "save"
	LOAD = "load"
)
const CONFIG = "config"

func init() {
	Index.MergeCommands(ice.Commands{
		CONFIG: {Name: "config key auto remove", Help: "配置", Actions: ice.Actions{
			SAVE: {Hand: func(m *ice.Message, arg ...string) { _config_save(m, arg[0], arg[1:]...) }},
			LOAD: {Hand: func(m *ice.Message, arg ...string) { _config_load(m, arg[0], arg[1:]...) }},
			mdb.LIST: {Hand: func(m *ice.Message, arg ...string) {
				list := []ice.Any{}
				for _, v := range arg[2:] {
					list = append(list, v)
				}
				m.Confv(arg[0], arg[1], list)
			}},
			mdb.REMOVE: {Name: "remove key sub", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.EXPORT, m.Option(mdb.KEY), m.Option(mdb.SUB), mdb.HASH, path.Join(ice.VAR_TRASH, kit.Keys(m.Option(mdb.KEY), m.Option(mdb.SUB))))
				nfs.Trash(m, path.Join(ice.VAR_DATA, m.Option(mdb.KEY)))
				m.Go(func() { m.Cmd(ice.EXIT, 1) })
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_config_list(m)
			} else {
				_config_make(m, arg[0], arg[1:]...)
				DisplayStoryJSON(m)
				m.Action(mdb.REMOVE)
			}
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
func ConfigAuto(m *ice.Message, arg ...string) {
	if cs := m.Target().Configs; cs[m.CommandKey()] == nil {
		cs[m.CommandKey()] = &ice.Config{Value: kit.Data()}
		ice.Info.Load(m, m.CommandKey())
	}
}
func ConfigFromOption(m *ice.Message, arg ...string) {
	for _, k := range arg {
		m.Config(k, kit.Select(m.Config(k), m.Option(k)))
	}
}
