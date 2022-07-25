package ice

import (
	"sync"

	kit "shylinux.com/x/toolkits"
)

var lock = map[string]*sync.RWMutex{}
var _lock = sync.Mutex{}

func (m *Message) _lock(key string) *sync.RWMutex {
	if key == "" {
		key = m.PrefixKey()
	}

	_lock.Lock()
	defer _lock.Unlock()

	l, ok := lock[key]
	if !ok {
		l = &sync.RWMutex{}
		lock[key] = l
	}
	return l
}
func (m *Message) Lock(arg ...Any) func() {
	l := m._lock(kit.Keys(arg...))
	l.Lock()
	return func() { l.Unlock() }
}
func (m *Message) RLock(arg ...Any) func() {
	l := m._lock(kit.Keys(arg...))
	l.RLock()
	return func() { l.RUnlock() }
}
