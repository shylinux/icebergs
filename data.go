package ice

import (
	"path"
	"strings"
	"sync"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/miss"
)

func (m *Message) ActionKey() string {
	return strings.TrimSuffix(strings.TrimPrefix(m._sub, PS), PS)
}
func (m *Message) CommandKey() string {
	return strings.TrimSuffix(strings.TrimPrefix(m._key, PS), PS)
}
func (m *Message) RoutePath(arg ...string) string {
	return m.Target().RoutePath(arg...)
}
func (m *Message) PrefixKey(arg ...string) string {
	return kit.Keys(m.Prefix(m.CommandKey()), arg)
}
func (m *Message) Prefix(arg ...string) string {
	return m.Target().PrefixKey(arg...)
}
func (m *Message) ConfigSet(keys string, arg ...string) {
	for i, k := range kit.Split(keys) {
		m.Config(k, kit.Select("", arg, i))
	}
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
func (m *Message) Configm(key string, arg ...Any) Map {
	v, _ := m.Configv(key, arg...).(Map)
	return v
}
func (m *Message) ConfigSimple(key ...string) (list []string) {
	for _, k := range kit.Split(kit.Join(key)) {
		list = append(list, k, m.Config(k))
	}
	return
}
func (m *Message) ConfigOption(key ...string) {
	for _, k := range kit.Split(kit.Join(key, FS)) {
		m.Config(k, kit.Select(m.Config(k), m.Option(k)))
	}
}
func (m *Message) Save(arg ...string) *Message {
	if !strings.Contains(Getenv("ctx_daemon"), "ctx") {
		return m
	}
	if len(arg) == 0 {
		for k := range m.target.Configs {
			arg = append(arg, k)
		}
	}
	for i, k := range arg {
		arg[i] = m.Prefix(k)
	}
	return m.Cmd(CONFIG, SAVE, m.Prefix(JSON), arg)
}
func (m *Message) Load(arg ...string) *Message {
	if !strings.Contains(Getenv("ctx_daemon"), "ctx") {
		return m
	}
	if len(arg) == 0 {
		for k := range m.target.Configs {
			arg = append(arg, k)
		}
	}
	for i, k := range arg {
		arg[i] = m.Prefix(k)
	}
	return m.Cmd(CONFIG, LOAD, m.Prefix(JSON), arg)
}

func (m *Message) Richs(prefix string, chain Any, raw Any, cb Any) (res Map) {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		return nil
	}

	switch cb := cb.(type) {
	case func(*sync.Mutex, string, Map):
		wg, mu := &sync.WaitGroup{}, &sync.Mutex{}
		defer wg.Wait()
		res = miss.Richs(path.Join(prefix, kit.Keys(chain)), cache, raw, func(key string, value Map) {
			wg.Add(1)
			m.Go(func() {
				defer wg.Done()
				cb(mu, key, value)
			})
		})
	default:
		res = miss.Richs(path.Join(prefix, kit.Keys(chain)), cache, raw, cb)
	}
	return res

}
func (m *Message) Rich(prefix string, chain Any, data Any) string {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		cache = kit.Data()
		m.Confv(prefix, chain, cache)
	}
	return miss.Rich(path.Join(prefix, kit.Keys(chain)), cache, data)
}
func (m *Message) Grow(prefix string, chain Any, data Any) int {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		cache = kit.Data()
		m.Confv(prefix, chain, cache)
	}
	return miss.Grow(path.Join(prefix, kit.Keys(chain)), cache, data)
}
func (m *Message) Grows(prefix string, chain Any, match string, value string, cb Any) Map {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		return nil
	}

	limit := kit.Int(m.Option(CACHE_LIMIT))
	if begin := kit.Int(m.Option(CACHE_BEGIN)); begin != 0 && limit > 0 {
		count := kit.Int(m.Option(CACHE_COUNT, kit.Int(kit.Value(cache, kit.Keym("count")))))
		if begin > 0 {
			m.Option(CACHE_OFFEND, count-begin-limit)
		} else {
			m.Option(CACHE_OFFEND, -begin-limit)
		}
	}

	return miss.Grows(path.Join(prefix, kit.Keys(chain)), cache,
		kit.Int(kit.Select("0", strings.TrimPrefix(m.Option(CACHE_OFFEND), "-"))),
		kit.Int(kit.Select("10", m.Option(CACHE_LIMIT))),
		match, value, cb)
}
