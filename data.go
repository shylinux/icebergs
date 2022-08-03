package ice

import (
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) ActionKey() string {
	return strings.TrimSuffix(strings.TrimPrefix(m._sub, PS), PS)
}
func (m *Message) CommandKey() string {
	return strings.TrimSuffix(strings.TrimPrefix(m._key, PS), PS)
}
func (m *Message) PrefixKey(arg ...string) string {
	return kit.Keys(m.Prefix(m.CommandKey()), arg)
}
func (m *Message) Prefix(arg ...string) string {
	return m.Target().PrefixKey(arg...)
}
func (m *Message) Config(key string, arg ...Any) string {
	if len(arg) > 0 {
		m.Conf(m.PrefixKey(), kit.Keym(key), arg[0])
	}
	return m.Conf(m.PrefixKey(), kit.Keym(key))
}
func (m *Message) Configv(key string, arg ...Any) Any {
	if len(arg) > 0 {
		m.Confv(m.PrefixKey(), kit.Keym(key), arg[0])
	}
	return m.Confv(m.PrefixKey(), kit.Keym(key))
}
func (m *Message) ConfigSimple(key ...string) (list []string) {
	for _, k := range kit.Split(kit.Join(key)) {
		list = append(list, k, m.Config(k))
	}
	return
}
