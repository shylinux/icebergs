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

func _config_format_list(m *ice.Message, v ice.Any) string {
	list := []string{"{", lex.NL}
	kit.For(v, func(k string, v ice.Any) {
		if k == mdb.HASH {
			list = append(list, lex.TB, kit.Format("%q", k), nfs.DF, "{", lex.NL)
			kit.For(v, func(k string, v ice.Any) {
				list = append(list, lex.TB, lex.TB, kit.Format("%q", k), nfs.DF, kit.Format(v), mdb.FS, lex.NL)
			})
			list = list[:len(list)-2]
			list = append(list, lex.TB, lex.NL, lex.TB, "}", mdb.FS, lex.NL)
		} else {
			list = append(list, lex.TB, kit.Format("%q", k), nfs.DF, kit.Format(v), mdb.FS, lex.NL)
		}
	})
	list = list[:len(list)-2]
	list = append(list, lex.NL, "}")
	return strings.Join(list, "")
}
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
			} else {
				for k := range v {
					if kit.IsIn(k, mdb.IMPORTANT) && len(v) > 1 {
						return false
					}
				}
			}
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
		kit.If(strings.HasPrefix(arg[1], mdb.AT), func() { arg[1] = msg.Cmdx(nfs.CAT, arg[1][1:]) })
		if !kit.IsIn(strings.Split(arg[0], nfs.PT)[0], mdb.META, mdb.HASH, mdb.LIST) {
			arg[0] = kit.Keys(mdb.META, arg[0])
		}
		mdb.Confv(msg, key, arg[0], kit.Parse(nil, "", arg[1:]...))
	}
	if len(arg) > 0 {
		m.Echo(FormatPretty(mdb.Confv(msg, key, arg[0]), 0, 1))
	} else {
		m.Echo(FormatPretty(mdb.Confv(msg, key), 0, 1))
		m.StatusTime(mdb.COUNT, kit.Length(mdb.Confv(msg, key, mdb.HASH)))
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
		CONFIG: {Name: "config key auto", Help: "配置", Actions: ice.Actions{
			SAVE:       {Hand: func(m *ice.Message, arg ...string) { _config_save(m, arg[0], arg[1:]...) }},
			LOAD:       {Hand: func(m *ice.Message, arg ...string) { _config_load(m, arg[0], arg[1:]...) }},
			mdb.EXPORT: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(mdb.EXPORT, arg[0], "", mdb.HASH) }},
			mdb.IMPORT: {Hand: func(m *ice.Message, arg ...string) { m.Cmdy(mdb.IMPORT, arg[0], "", mdb.HASH) }},
			nfs.TRASH: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.EXPORT, arg[0], "", mdb.HASH, path.Join(ice.VAR_TRASH, kit.Keys(arg[0])))
				nfs.Trash(m, path.Join(ice.VAR_DATA, arg[0]))
				mdb.Config(m, arg[0], nil, nil)
				m.Go(func() { m.Cmd(ice.EXIT, 1) })
			}},
			mdb.LIST: {Hand: func(m *ice.Message, arg ...string) {
				list := []ice.Any{}
				kit.For(arg[2:], func(v string) { list = append(list, v) })
				mdb.Confv(m, arg[0], arg[1], list)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				_config_list(m)
			} else {
				_config_make(m, arg[0], arg[1:]...)
				m.Action(mdb.EXPORT, mdb.IMPORT, nfs.TRASH)
			}
		}},
	})
}
func init() { ice.Info.Save = Save; ice.Info.Load = Load }
func Save(m *ice.Message, arg ...string) *ice.Message {
	kit.If(len(arg) == 0, func() { arg = kit.SortedKey(m.Target().Configs) })
	kit.For(arg, func(i int, k string) { arg[i] = strings.Replace(m.Prefix(k), nfs.PS, "", 1) })
	return m.Cmd(CONFIG, SAVE, m.Prefix(nfs.JSON), arg)
}
func Load(m *ice.Message, arg ...string) *ice.Message {
	kit.If(len(arg) == 0, func() { arg = kit.SortedKey(m.Target().Configs) })
	kit.For(arg, func(i int, k string) { arg[i] = strings.Replace(m.Prefix(k), nfs.PS, "", 1) })
	return m.Cmd(CONFIG, LOAD, m.Prefix(nfs.JSON), arg)
}
func ConfAction(arg ...ice.Any) ice.Actions { return ice.Actions{ice.CTX_INIT: mdb.AutoConfig(arg...)} }
func ConfigSimple(m *ice.Message, key ...string) (res []string) {
	kit.For(kit.Split(kit.Join(key)), func(k string) { res = append(res, k, mdb.Config(m, k)) })
	return
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
func FormatPretty(v ice.Any, i, n int) string {
	return kit.Formats(v)
	switch v := v.(type) {
	case map[string]ice.Any:
		if n == 0 {
			list := []string{"{"}
			kit.For(v, func(k string, v ice.Any) {
				list = append(list, kit.Format("%q", k), nfs.DF, FormatPretty(v, 0, 0), mdb.FS)
			})
			list = list[:len(list)-1]
			list = append(list, "}")
			return strings.Join(list, "")
		}
		list := []string{"{", lex.NL}
		kit.For(v, func(k string, v ice.Any) {
			list = append(list, strings.Repeat(lex.TB, i+1), kit.Format("%q", k), nfs.DF)
			if i < n && !kit.IsIn(k, mdb.META) && !strings.HasPrefix(k, "_") {
				list = append(list, FormatPretty(v, i+1, n))
			} else {
				list = append(list, FormatPretty(v, 0, 0))
			}
			list = append(list, mdb.FS, lex.NL)
		})
		list = append(list[:len(list)-2], lex.NL)
		list = append(list, strings.Repeat(lex.TB, i), "}")
		return strings.Join(list, "")
	case []ice.Any:
		if n == 0 {
			list := []string{"["}
			kit.For(v, func(k string, v ice.Any) {
				list = append(list, FormatPretty(v, 0, 0), mdb.FS)
			})
			list = list[:len(list)-1]
			list = append(list, "]")
			return strings.Join(list, "")
		}
		list := []string{"[", lex.NL}
		kit.For(v, func(v ice.Any) {
			list = append(list, strings.Repeat(lex.TB, i+1))
			if i < n {
				list = append(list, FormatPretty(v, i+1, n))
			} else {
				list = append(list, FormatPretty(v, 0, 0))
			}
			list = append(list, mdb.FS, lex.NL)
		})
		list = append(list[:len(list)-2], lex.NL)
		list = append(list, strings.Repeat(lex.TB, i), "]")
		return strings.Join(list, "")
	case string:
		return kit.Format(v)
		return kit.Format("%q", v)
	default:
		return kit.Format(v)
	}
}
