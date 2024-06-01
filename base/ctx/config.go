package ctx

import (
	"encoding/json"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
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
		for k, v := range v {
			if v, ok := v.(ice.Map); ok && len(v) == 0 {
				continue
			} else if kit.IndexOf(arg, k) == -1 {
				return false
			}
		}
		return true
	}
	return false
}
func _config_save(m *ice.Message, name string, arg ...string) {
	if !ice.HasVar() {
		return
	}
	data, msg := ice.Map{}, m.Spawn(m.Source())
	for _, k := range arg {
		if v := mdb.Confv(msg, k); _config_only(v, mdb.META) && _config_only(kit.Value(v, mdb.META),
			mdb.IMPORTANT, mdb.EXPIRE, mdb.VENDOR, nfs.SOURCE, nfs.SCRIPT, nfs.PATH, lex.REGEXP,
			mdb.SHORT, mdb.FIELD, mdb.SHORTS, mdb.FIELDS,
			mdb.ACTION, mdb.SORT, mdb.TOOLS,
			"link", "linux", "darwin", "windows",
		) {
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
		if s, e := json.MarshalIndent(data, "", "  "); !m.WarnNotValid(e) {
			if _, e := f.Write(s); !m.WarnNotValid(e) {
			}
		}
	}
}
func _config_load(m *ice.Message, name string, arg ...string) {
	if !ice.HasVar() {
		return
	}
	if f, e := miss.OpenFile(path.Join(ice.VAR_CONF, name)); e == nil {
		defer f.Close()
		data, msg := ice.Map{}, m.Spawn(m.Source())
		json.NewDecoder(f).Decode(&data)
		for k, v := range data {
			msg.Search(k, func(p *ice.Context, s *ice.Context, key string, conf *ice.Config) {
				kit.If(s.Configs[key] == nil, func() { s.Configs[key] = &ice.Config{} })
				s.Configs[key].Value = v
			})
		}
	}
}
func _config_make(m *ice.Message, key string, arg ...string) {
	msg := m.Spawn(m.Source())
	if len(arg) > 1 {
		kit.If(!kit.IsIn(strings.Split(arg[0], nfs.PT)[0], mdb.META, mdb.HASH, mdb.LIST), func() { arg[0] = kit.Keys(mdb.META, arg[0]) })
		kit.If(strings.HasPrefix(arg[1], mdb.AT), func() { arg[1] = msg.Cmdx(nfs.CAT, arg[1][1:]) })
		mdb.Confv(msg, key, arg[0], kit.Parse(nil, "", arg[1:]...))
	}
	if len(arg) > 0 {
		m.Echo(kit.Formats(mdb.Confv(msg, key, arg[0])))
	} else {
		m.Echo(kit.Formats(mdb.Confv(msg, key))).StatusTime(mdb.COUNT, kit.Length(mdb.Confv(msg, key, mdb.HASH)))
	}
}
func _config_list(m *ice.Message) {
	for k, v := range m.Source().Configs {
		if !IsOrderCmd(k) {
			m.Push(mdb.KEY, k).Push(mdb.NAME, v.Name).Push(mdb.VALUE, kit.Format(v.Value))
		}
	}
	m.Sort(mdb.KEY)
}

const CONFIG = "config"

func init() {
	Index.MergeCommands(ice.Commands{
		CONFIG: {Name: "config key auto", Help: "配置", Actions: ice.Actions{
			nfs.SAVE:   {Hand: func(m *ice.Message, arg ...string) { _config_save(m, arg[0], arg[1:]...) }},
			nfs.LOAD:   {Hand: func(m *ice.Message, arg ...string) { _config_load(m, arg[0], arg[1:]...) }},
			mdb.EXPORT: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(arg[0], mdb.EXPORT) }},
			mdb.IMPORT: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(arg[0], mdb.IMPORT) }},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				nfs.Trash(m, path.Join(ice.VAR_DATA, arg[0]))
				nfs.Trash(m, m.Cmdx(arg[0], mdb.EXPORT))
				mdb.Conf(m, arg[0], mdb.HASH, "")
			}},
			mdb.CREATE: {Name: "create name value", Hand: func(m *ice.Message, arg ...string) {
				m.Confv(m.Option(mdb.KEY), kit.Keys(mdb.META, m.Option(mdb.NAME)), m.Option(mdb.VALUE))
			}},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) {
				m.Confv(m.Option(mdb.KEY), kit.Keys(mdb.META, m.Option(mdb.NAME)), "")
			}},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.VALUE {
					m.Confv(m.Option(mdb.KEY), kit.Keys(mdb.META, m.Option(mdb.NAME)), arg[1])
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_config_list(m)
			} else {
				_config_make(m, arg[0], arg[1:]...)
				m.Action(mdb.CREATE, mdb.IMPORT, mdb.EXPORT, nfs.TRASH)
				kit.For(mdb.Confv(m, arg[0], mdb.META), func(k, v string) {
					m.Push(mdb.NAME, k).Push(mdb.VALUE, v).PushButton(mdb.REMOVE)
				})
				DisplayStoryJSON(m)
			}
		}},
	})
}
func init() { ice.Info.Save = Save; ice.Info.Load = Load }
func Save(m *ice.Message, arg ...string) *ice.Message {
	kit.If(len(arg) == 0, func() { arg = kit.SortedKey(m.Target().Configs) })
	kit.For(arg, func(i int, k string) { arg[i] = strings.Replace(m.Prefix(k), nfs.PS, "", 1) })
	return m.Cmd(prefix(CONFIG), nfs.SAVE, m.Prefix(nfs.JSON), arg)
}
func Load(m *ice.Message, arg ...string) *ice.Message {
	kit.If(len(arg) == 0, func() { arg = kit.SortedKey(m.Target().Configs) })
	kit.For(arg, func(i int, k string) { arg[i] = strings.Replace(m.Prefix(k), nfs.PS, "", 1) })
	return m.Cmd(prefix(CONFIG), nfs.LOAD, m.Prefix(nfs.JSON), arg)
}
func ConfAction(arg ...ice.Any) ice.Actions {
	return ice.Actions{ice.CTX_INIT: mdb.AutoConfig(arg...)}
}
func ConfigFromOption(m *ice.Message, arg ...string) {
	if len(arg) == 0 {
		kit.For(m.Target().Commands[m.CommandKey()].Actions[m.ActionKey()].List, func(value ice.Any) {
			arg = append(arg, kit.Format(kit.Value(value, mdb.NAME)))
		})
	}
	kit.For(arg, func(k string) { mdb.Config(m, k, kit.Select(mdb.Config(m, k), m.Option(k))) })
}
func OptionFromConfig(m *ice.Message, arg ...string) string {
	kit.For(arg, func(k string) { m.Option(k, mdb.Config(m, k)) })
	return m.Option(arg[0])
}
