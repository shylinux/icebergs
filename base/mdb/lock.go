package mdb

import (
	"sync"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/task"
)

var _lock = task.Lock{}
var _locks = map[string]*task.Lock{}

func getLock(m *ice.Message, arg ...string) *task.Lock {
	key := kit.Select(m.PrefixKey(), kit.Keys(arg))
	defer _lock.Lock()()
	l, ok := _locks[key]
	kit.If(!ok, func() { l = &task.Lock{}; _locks[key] = l })
	return l
}
func Lock(m *ice.Message, arg ...string) func() {
	return getLock(m, arg...).Lock()
}
func RLock(m *ice.Message, arg ...string) func() {
	return getLock(m, arg...).RLock()
}

func ConfigSimple(m *ice.Message, key ...string) (res []string) {
	for _, key := range key {
		res = append(res, key, kit.Format(Configv(m, key)))
	}
	return
}
func Config(m *ice.Message, key string, arg ...Any) string {
	return kit.Format(Configv(m, key, arg...))
}
func Configv(m *ice.Message, key string, arg ...Any) Any {
	kit.If(len(arg) > 0, func() { Confv(m, m.PrefixKey(), kit.Keym(key), arg[0]) })
	return Confv(m, m.PrefixKey(), kit.Keym(key))
}
func Confv(m *ice.Message, arg ...Any) Any {
	key := kit.Select(m.PrefixKey(), kit.Format(arg[0]))
	if ctx, ok := ice.Info.Index[key].(*ice.Context); ok {
		key = ctx.Prefix(key)
	}
	if len(arg) > 2 {
		defer Lock(m, key)()
	} else {
		defer RLock(m, key)()
	}
	return m.Confv(arg...)
}
func Conf(m *ice.Message, arg ...Any) string { return kit.Format(Confv(m, arg...)) }
func Confm(m *ice.Message, key string, sub Any, cbs ...Any) Map {
	val := m.Confv(key, sub)
	kit.If(len(cbs) > 0, func() { kit.For(val, cbs[0]) })
	value, _ := val.(Map)
	return value
}

var cache = sync.Map{}

func Cache(m *ice.Message, key string, add func() Any) Any {
	if key = kit.Keys(m.PrefixKey(), key); add == nil {
		cache.Delete(key)
		return nil
	} else if val, ok := cache.Load(key); ok {
		return val
	} else if val := add(); val != nil {
		cache.Store(key, val)
		return val
	} else {
		return nil
	}
}
