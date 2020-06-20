package ice

import (
	kit "github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/miss"
)

func (m *Message) Richs(key string, chain interface{}, raw interface{}, cb interface{}) (res map[string]interface{}) {
	cache := m.Confm(key, chain)
	if cache == nil {
		return nil
	}
	return miss.Richs(kit.Keys(key, chain), cache, raw, cb)
}
func (m *Message) Rich(key string, chain interface{}, data interface{}) string {
	cache := m.Confm(key, chain)
	if cache == nil {
		cache = map[string]interface{}{}
		m.Confv(key, chain, cache)
	}
	return miss.Rich(kit.Keys(key, chain), cache, data)
}
func (m *Message) Grow(key string, chain interface{}, data interface{}) int {
	cache := m.Confm(key, chain)
	if cache == nil {
		cache = map[string]interface{}{}
		m.Confv(key, chain, cache)
	}
	return miss.Grow(kit.Keys(key, chain), cache, data)
}
func (m *Message) Grows(key string, chain interface{}, match string, value string, cb interface{}) map[string]interface{} {
	cache := m.Confm(key, chain)
	if cache == nil {
		return nil
	}
	return miss.Grows(kit.Keys(key, chain), cache,
		kit.Int(kit.Select("0", m.Option("cache.offend"))),
		kit.Int(kit.Select("10", m.Option("cache.limit"))),
		match, value, cb)
}
