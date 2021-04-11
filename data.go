package ice

import (
	"sync"

	kit "github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/miss"
)

func (m *Message) Richs(prefix string, chain interface{}, raw interface{}, cb interface{}) (res map[string]interface{}) {
	cache := m.Confm(prefix, chain)
	if cache == nil {
		return nil
	}

	switch cb := cb.(type) {
	case func(*sync.Mutex, string, map[string]interface{}):
		mu := &sync.Mutex{}
		wg := &sync.WaitGroup{}
		res = miss.Richs(kit.Keys(prefix, chain), cache, raw, func(key string, value map[string]interface{}) {
			wg.Add(1)
			m.Go(func() {
				defer wg.Done()
				cb(mu, key, value)
			})
		})
		wg.Wait()
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

	limit := kit.Int(m.Option("cache.limit"))
	if begin := kit.Int(m.Option("cache.begin")); begin != 0 && limit > 0 {
		count := kit.Int(m.Option("cache.count", kit.Int(kit.Value(cache, "meta.count"))))
		if begin > 0 {
			m.Option("cache.offend", count-begin-limit)
		} else {
			m.Option("cache.offend", -begin-limit)
		}
	}

	return miss.Grows(kit.Keys(prefix, chain), cache,
		kit.Int(kit.Select("0", m.Option("cache.offend"))),
		kit.Int(kit.Select("10", m.Option("cache.limit"))),
		match, value, cb)
}
