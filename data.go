package ice

import (
	"os"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) ActionKey() string           { return strings.TrimPrefix(strings.TrimSuffix(m._sub, PS), PS) }
func (m *Message) CommandKey() string          { return strings.TrimPrefix(strings.TrimSuffix(m._key, PS), PS) }
func (m *Message) PrefixKey() string           { return m.Prefix(m.CommandKey()) }
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

func SaveImportant(m *Message, arg ...string) {
	if Info.Important != true {
		return
	}
	kit.For(arg, func(i int, v string) { kit.If(v == "" || strings.Contains(v, SP), func() { arg[i] = "\"" + v + "\"" }) })
	m.Cmd("nfs.push", VAR_DATA_IMPORTANT, kit.Join(arg, SP), NL)
}
func loadImportant(m *Message) {
	if f, e := os.Open(VAR_DATA_IMPORTANT); e == nil {
		defer f.Close()
		kit.For(f, func(s string) { kit.If(s != "" && !strings.HasPrefix(s, "# "), func() { m.Cmd(kit.Split(s)) }) })
	}
	Info.Important = true
}
func removeImportant(m *Message) { os.Remove(VAR_DATA_IMPORTANT) }
