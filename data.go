package ice

import (
	"strings"
	"sync"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/miss"
)

func (m *Message) CommandKey() string {
	return strings.TrimSuffix(strings.TrimPrefix(m._key, PS), PS)
}
func (m *Message) PrefixKey(arg ...string) string {
	return kit.Keys(m.Cap(CTX_FOLLOW), m.CommandKey(), arg)
}
func (m *Message) Prefix(arg ...string) string {
	return kit.Keys(m.Cap(CTX_FOLLOW), arg)
}
func (m *Message) Config(key string, arg ...interface{}) string {
	if len(arg) > 0 {
		m.Conf(m.PrefixKey(), kit.Keym(key), arg[0])
	}
	return m.Conf(m.PrefixKey(), kit.Keym(key))
	// return kit.Format(m.Configv(key, arg...))
}
func (m *Message) Configv(key string, arg ...interface{}) interface{} {
	if len(arg) > 0 {
		m.Confv(m.PrefixKey(), kit.Keym(key), arg[0])
	}
	return m.Confv(m.PrefixKey(), kit.Keym(key))
}
func (m *Message) Configm(key string, arg ...interface{}) map[string]interface{} {
	v, _ := m.Configv(key, arg...).(map[string]interface{})
	return v
}
func (m *Message) ConfigSimple(key ...string) (list []string) {
	for _, k := range kit.Split(kit.Join(key)) {
		list = append(list, k, m.Config(k))
	}
	return
}
func (m *Message) Save(arg ...string) *Message {
	if len(arg) == 0 {
		for k := range m.target.Configs {
			arg = append(arg, k)
		}
	}
	for i, k := range arg {
		arg[i] = m.Prefix(k)
	}
	return m.Cmd("config", SAVE, m.Prefix("json"), arg)
}
func (m *Message) Load(arg ...string) *Message {
	if len(arg) == 0 {
		for k := range m.target.Configs {
			arg = append(arg, k)
		}
	}
	for i, k := range arg {
		arg[i] = m.Prefix(k)
	}
	return m.Cmd("config", LOAD, m.Prefix("json"), arg)
}

func (m *Message) Richs(prefix string, chain interface{}, raw interface{}, cb interface{}) (res map[string]interface{}) {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		return nil
	}

	switch cb := cb.(type) {
	case func(*sync.Mutex, string, map[string]interface{}):
		wg, mu := &sync.WaitGroup{}, &sync.Mutex{}
		defer wg.Wait()
		res = miss.Richs(kit.Keys(prefix, chain), cache, raw, func(key string, value map[string]interface{}) {
			wg.Add(1)
			m.Go(func() {
				defer wg.Done()
				cb(mu, key, value)
			})
		})
	default:
		res = miss.Richs(kit.Keys(prefix, chain), cache, raw, cb)
	}
	return res

}
func (m *Message) Rich(prefix string, chain interface{}, data interface{}) string {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		cache = kit.Data()
		m.Confv(prefix, chain, cache)
	}
	return miss.Rich(kit.Keys(prefix, chain), cache, data)
}
func (m *Message) Grow(prefix string, chain interface{}, data interface{}) int {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		cache = kit.Data()
		m.Confv(prefix, chain, cache)
	}
	return miss.Grow(kit.Keys(prefix, chain), cache, data)
}
func (m *Message) Grows(prefix string, chain interface{}, match string, value string, cb interface{}) map[string]interface{} {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		return nil
	}

	limit := kit.Int(m.Option(CACHE_LIMIT))
	if begin := kit.Int(m.Option(CACHE_BEGIN)); begin != 0 && limit > 0 {
		count := kit.Int(m.Option(CACHE_COUNT, kit.Int(kit.Value(cache, kit.Keym(kit.MDB_COUNT)))))
		if begin > 0 {
			m.Option(CACHE_OFFEND, count-begin-limit)
		} else {
			m.Option(CACHE_OFFEND, -begin-limit)
		}
	}

	return miss.Grows(kit.Keys(prefix, chain), cache,
		kit.Int(kit.Select("0", strings.TrimPrefix(m.Option(CACHE_OFFEND), "-"))),
		kit.Int(kit.Select("10", m.Option(CACHE_LIMIT))),
		match, value, cb)
}
