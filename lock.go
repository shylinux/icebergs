package ice

import (
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) index(key string, index int, value ...string) string {
	if len(value) > 0 {
		defer m.lock.Lock()()
		m._meta[key][index] = value[0]
	} else {
		defer m.lock.RLock()()
	}
	return kit.Select("", m._meta[key], index)
}
func (m *Message) value(key string, list ...string) []string {
	if len(list) > 0 {
		defer m.lock.Lock()()
		m._meta[key] = list
	} else {
		defer m.lock.RLock()()
	}
	return m._meta[key]
}
func (m *Message) delete(key ...string) {
	defer m.lock.Lock()()
	for _, key := range key {
		delete(m._meta, key)
	}
}

func (m *Message) Add(key string, arg ...string) *Message {
	if len(arg) == 0 {
		return m
	}
	defer m.lock.Lock()()
	switch key {
	case MSG_DETAIL, MSG_RESULT:
		m._meta[key] = append(m._meta[key], arg...)
	case MSG_OPTION, MSG_APPEND:
		if index := 0; key == MSG_APPEND {
			if m._meta[MSG_OPTION], index = kit.SliceRemove(m._meta[MSG_OPTION], arg[0]); index > -1 {
				delete(m._meta, arg[0])
			}
		}
		if m._meta[arg[0]] = append(m._meta[arg[0]], arg[1:]...); kit.IndexOf(m._meta[key], arg[0]) == -1 {
			m._meta[key] = append(m._meta[key], arg[0])
		}
	}
	return m
}
func (m *Message) setDetail(key string, arg ...string) *Message {
	defer m.lock.Lock()()
	for i := 0; i < len(m._meta[KEY]); i++ {
		if m._meta[KEY][i] == key {
			if len(arg) > 0 {
				m._meta[VALUE][i] = arg[0]
				return m
			}
			for ; i < len(m._meta[KEY])-1; i++ {
				m._meta[KEY][i] = m._meta[KEY][i+1]
				m._meta[VALUE][i] = m._meta[VALUE][i+1]
			}
			m._meta[KEY] = m._meta[KEY][0 : len(m._meta[KEY])-1]
			m._meta[VALUE] = m._meta[VALUE][0 : len(m._meta[VALUE])-1]
			return m
		}
	}
	if len(arg) > 0 {
		m._meta[KEY] = append(m._meta[KEY], key)
		m._meta[VALUE] = append(m._meta[VALUE], arg[0])
	}
	return m
}
func (m *Message) Optionv(key string, arg ...Any) Any {
	key = kit.Select(MSG_OPTION, key)
	key = strings.ReplaceAll(key, "*", "")
	var unlock func()
	if len(arg) > 0 {
		unlock = m.lock.Lock()
		kit.If(kit.IndexOf(m._meta[MSG_OPTION], key) == -1, func() { m._meta[MSG_OPTION] = append(m._meta[MSG_OPTION], key) })
		switch delete(m._data, key); v := arg[0].(type) {
		case nil:
			delete(m._meta, key)
		case string:
			func() {
				for i := 0; i < len(arg); i++ {
					if _, ok := arg[i].(string); !ok {
						m._data[key] = arg
						return
					}
				}
				m._meta[key] = kit.Simple(arg...)
			}()
		case []string:
			m._meta[key] = v
		default:
			if len(arg) > 1 {
				m._data[key] = arg
			} else {
				m._data[key] = v
			}
		}
	} else {
		unlock = m.lock.RLock()
	}
	if v, ok := m._data[key]; ok {
		unlock()
		return v
	} else if v, ok := m._meta[key]; ok {
		unlock()
		return v
	} else {
		unlock()
	}
	if m.message != nil {
		return m.message.Optionv(key)
	} else {
		return nil
	}
}
