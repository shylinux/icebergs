package ice

import (
	"sync"

	kit "shylinux.com/x/toolkits"
)

var lock = map[string]*sync.RWMutex{}
var _lock = sync.Mutex{}

func _get_lock(key string) (*sync.RWMutex, string) {
	_lock.Lock()
	defer _lock.Unlock()

	l, ok := lock[key]
	if !ok {
		l = &sync.RWMutex{}
		lock[key] = l
	}
	return l, key
}
func (m *Message) Lock(arg ...Any) func() {
	l, key := _get_lock(kit.Keys(arg...))
	m.Debug("before lock %v", key)
	l.Lock()
	m.Debug("success lock %v", key)
	return func() {
		l.Unlock()
		m.Debug("success unlock %v", key)
	}
}
func (m *Message) RLock(arg ...Any) func() {
	l, key := _get_lock(kit.Keys(arg...))
	m.Debug("before rlock %v", key)
	l.RLock()
	m.Debug("success rlock %v", key)
	return func() {
		l.RUnlock()
		m.Debug("success runlock %v", key)
	}
}
