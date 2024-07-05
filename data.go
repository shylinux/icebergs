package ice

import (
	"os"
	"strings"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/task"
)

func (m *Message) ActionKey() string {
	action := strings.TrimPrefix(strings.TrimSuffix(m._sub, PS), PS)
	return kit.Select("", action, !kit.IsIn(action, LIST, SELECT))
}
func (m *Message) CommandKey() string { return strings.TrimPrefix(strings.TrimSuffix(m._key, PS), PS) }
func (m *Message) PrefixKey() string  { return m.Prefix(m.CommandKey()) }
func (m *Message) ShortKey() string {
	key := m.CommandKey()
	if p, ok := Info.Index[key].(*Context); ok && p == m.target {
		return key
	}
	return m.PrefixKey()
}
func (m *Message) Prefix(arg ...string) string { return m.Target().Prefix(arg...) }
func (m *Message) Confv(arg ...Any) (val Any) { // key sub value
	run := func(conf *Config) {
		if len(arg) == 1 {
			val = conf.Value
			return
		} else if len(arg) > 2 {
			if arg[1] == nil || arg[1] == "" {
				conf.Value = arg[2]
			} else {
				kit.Value(conf.Value, arg[1:]...)
			}
		}
		val = kit.Value(conf.Value, arg[1])
	}
	key := kit.Format(arg[0])
	kit.If(key == "", func() { key = m._key })
	if conf, ok := m.target.Configs[key]; ok {
		run(conf)
	} else if conf, ok := m.source.Configs[key]; ok {
		run(conf)
	} else {
		m.Search(key, func(p *Context, s *Context, key string, conf *Config) { run(conf) })
	}
	return
}
func (m *Message) Conf(arg ...Any) string { return kit.Format(m.Confv(arg...)) }

var _important = task.Lock{}

func SaveImportant(m *Message, arg ...string) {
	if !HasVar() {
		return
	}
	if Info.Important != true || len(arg) == 0 {
		return
	}
	kit.For(arg, func(v string, i int) { arg[i] = kit.Format("%q", v) })
	defer _important.Lock()()
	m.Cmd("nfs.push", VAR_DATA_IMPORTANT, kit.Join(arg, SP), NL)
}
func loadImportant(m *Message) {
	if !HasVar() {
		return
	}
	if f, e := os.Open(VAR_DATA_IMPORTANT); e == nil {
		defer f.Close()
		kit.For(f, func(s string) { kit.If(s != "" && !strings.HasPrefix(s, "# "), func() { m.Cmd(kit.Split(s)) }) })
	}
	// Info.Important = HasVar()
}
func removeImportant(m *Message) { os.Remove(VAR_DATA_IMPORTANT) }
