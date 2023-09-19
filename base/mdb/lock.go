package mdb

import (
	"sync"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/task"
)

type configMessage interface {
	Option(key string, arg ...Any) string
	PrefixKey() string
	Confv(...Any) Any
}

var _lock = task.Lock{}
var _locks = map[string]*task.Lock{}

func getLock(m configMessage, arg ...string) *task.Lock {
	key := kit.Select(m.PrefixKey(), kit.Keys(arg))
	defer _lock.Lock()()
	l, ok := _locks[key]
	kit.If(!ok, func() { l = &task.Lock{}; _locks[key] = l })
	return l
}
func Lock(m configMessage, arg ...string) func()  { return getLock(m, arg...).Lock() }
func RLock(m configMessage, arg ...string) func() { return getLock(m, arg...).RLock() }

func Config(m configMessage, key string, arg ...Any) string {
	return kit.Format(Configv(m, key, arg...))
}
func Configv(m configMessage, key string, arg ...Any) Any {
	kit.If(len(arg) > 0, func() { Confv(m, m.PrefixKey(), kit.Keym(key), arg[0]) })
	return Confv(m, m.PrefixKey(), kit.Keym(key))
}
func Confv(m configMessage, arg ...Any) Any {
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
func Conf(m configMessage, arg ...Any) string {
	return kit.Format(Confv(m, arg...))
}
func Confm(m configMessage, key string, sub Any, cbs ...Any) Map {
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
	}
	if val, ok := cache.Load(key); ok {
		return val
	}
	if val := add(); val != nil {
		cache.Store(key, val)
		return val
	}
	return nil
}
