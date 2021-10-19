package ice

import (
	"strings"
	"sync"

	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/miss"
)

func (m *Message) CommandKey(arg ...string) string {
	return strings.TrimPrefix(m._key, "/")
}
func (m *Message) PrefixKey(arg ...string) string {
	return kit.Keys(m.Cap(CTX_FOLLOW), strings.TrimPrefix(m._key, "/"), arg)
}
func (m *Message) Prefix(arg ...string) string {
	return kit.Keys(m.Cap(CTX_FOLLOW), arg)
}
func (m *Message) Config(key string) string {
	return m.Conf(m.PrefixKey(), kit.Keym(key))
}
func (m *Message) ConfigSimple(key string) []string {
	return []string{key, m.Conf(m.PrefixKey(), kit.Keym(key))}
}
func (m *Message) Save(arg ...string) *Message {
	if len(arg) == 0 {
		for k := range m.target.Configs {
			arg = append(arg, k)
		}
	}
	list := []string{}
	for _, k := range arg {
		list = append(list, m.Prefix(k))
	}
	m.Cmd("ctx.config", SAVE, m.Prefix("json"), list)
	return m
}
func (m *Message) Load(arg ...string) *Message {
	list := []string{}
	for _, k := range arg {
		list = append(list, m.Prefix(k))
	}
	m.Cmd("ctx.config", LOAD, m.Prefix("json"), list)
	return m
}

func (m *Message) Richs(prefix string, chain interface{}, raw interface{}, cb interface{}) (res map[string]interface{}) {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		return nil
	}

	switch cb := cb.(type) {
	case func(*sync.Mutex, string, map[string]interface{}):
		mu := &sync.Mutex{}

		wg := &sync.WaitGroup{}
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
