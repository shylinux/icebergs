package ice

import (
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) ActionKey() string {
	return strings.TrimPrefix(strings.TrimSuffix(m._sub, PS), PS)
}
func (m *Message) CommandKey() string {
	return strings.TrimPrefix(strings.TrimSuffix(m._key, PS), PS)
}
func (m *Message) PrefixKey(arg ...Any) string {
	return kit.Keys(m.Prefix(m.CommandKey()), kit.Keys(arg...))
}
func (m *Message) Prefix(arg ...string) string {
	return m.Target().PrefixKey(arg...)
}
func (m *Message) Config(key string, arg ...Any) string {
	return kit.Format(m.Configv(key, arg...))
}
func (m *Message) Configv(key string, arg ...Any) Any {
	if len(arg) > 0 {
		m.Confv(m.PrefixKey(), kit.Keym(key), arg[0])
	}
	return m.Confv(m.PrefixKey(), kit.Keym(key))
}
func (m *Message) ConfigSimple(key ...string) (res []string) {
	for _, k := range kit.Split(kit.Join(key)) {
		res = append(res, k, m.Config(k))
	}
	return
}
