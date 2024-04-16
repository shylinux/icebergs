package ice

import (
	"strconv"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) Set(key string, arg ...string) *Message {
	switch key {
	case MSG_DETAIL, MSG_RESULT:
		m.delete(key)
	case MSG_OPTION, MSG_APPEND:
		if m.FieldsIsDetail() {
			if len(arg) > 0 {
				m.setDetail(arg[0], arg[1:]...)
			} else {
				m.delete(KEY, VALUE, MSG_APPEND)
			}
		} else if len(arg) > 0 {
			if m.delete(arg[0]); len(arg) > 1 {
				m.value(arg[0], arg[1:]...)
			} else {
				list := m.value(key)
				for i, k := range list {
					if k == arg[0] {
						for ; i < len(list)-1; i++ {
							list[i] = list[i+1]
						}
						list = list[:len(list)-1]
						m.value(key, list...)
						break
					}
				}
			}
		} else {
			kit.For(m.value(key), func(k string) { m.delete(k) })
			m.delete(key)
		}
		return m
	default:
		if m.FieldsIsDetail() {
			return m.setDetail(key, arg...)
		}
		kit.For(kit.Split(key), func(k string) { m.delete(k) })
	}
	if len(arg) == 0 {
		return m
	}
	return m.Add(key, arg...)
}
func (m *Message) Cut(fields ...string) *Message {
	m.value(MSG_APPEND, kit.Split(kit.Join(fields))...)
	return m
}
func (m *Message) CutTo(key, to string) *Message {
	return m.Cut(key).RenameAppend(key, to)
}
func (m *Message) Push(key string, value Any, arg ...Any) *Message {
	head := kit.Simple()
	kit.If(len(head) == 0 && len(arg) > 0, func() { head = kit.Simple(arg[0]) })
	kit.If(len(head) == 0, func() { head = kit.Simple(m.value(MSG_APPEND)) })
	kit.If(len(head) == 0 && !m.FieldsIsDetail(), func() { head = kit.Split(m.OptionFields()) })
	var val Map
	kit.If(len(arg) > 1, func() {
		switch v := arg[1].(type) {
		case Map:
			val = v
		default:
			val = kit.Dict(v)
		}
	})
	switch value := value.(type) {
	case Map:
		kit.If(len(head) == 0, func() { head = kit.SortedKey(kit.KeyValue(nil, "", value)) })
		kit.For(head, func(k string) {
			k = strings.TrimSuffix(k, "*")
			var v Any
			switch k {
			case "_target":
				return
			case KEY, HASH:
				if key != "" && key != FIELDS_DETAIL {
					v = key
					break
				}
				fallthrough
			default:
				if v = value[k]; v != nil {
					break
				}
				if v = kit.Value(value, k); v != nil {
					break
				}
				if v = kit.Value(value, kit.Keys(EXTRA, k)); v != nil {
					break
				}
				if v = val[k]; v != nil {
					break
				}
				if v = kit.Value(val, k); v != nil {
					break
				}
				if v = kit.Value(val, kit.Keys(EXTRA, k)); v != nil {
					break
				}
			}
			switch v := kit.Format(v); key {
			case FIELDS_DETAIL:
				m.Add(MSG_APPEND, KEY, strings.TrimPrefix(k, EXTRA+PT)).Add(MSG_APPEND, VALUE, v)
			default:
				m.Add(MSG_APPEND, k, v)
			}
		})
	case Maps:
		kit.If(len(head) == 0, func() { head = kit.SortedKey(value) })
		kit.For(head, func(k string) {
			k = strings.TrimSuffix(k, "*")
			m.Push(k, kit.Select(kit.Format(val[k]), value[k]))
		})
	default:
		kit.For(kit.Simple(value, arg), func(v string) {
			key = strings.TrimSuffix(key, "*")
			if m.FieldsIsDetail() {
				m.Add(MSG_APPEND, KEY, key).Add(MSG_APPEND, VALUE, kit.Format(value))
			} else {
				m.Add(MSG_APPEND, key, v)
			}
		})
	}
	return m
}
func (m *Message) EchoLine(str string, arg ...Any) *Message {
	return m.Echo(str, arg...).Echo(NL)
}
func (m *Message) Echo(str string, arg ...Any) *Message {
	if str == "" {
		return m
	}
	return m.Add(MSG_RESULT, kit.Format(str, arg...))
}
func (m *Message) Copy(msg *Message, arg ...string) *Message {
	if m == nil || msg == nil || m == msg {
		return m
	}
	if len(arg) > 0 {
		kit.For(arg[1:], func(k string) { m.Add(arg[0], kit.Simple(k, msg.value(k))...) })
		return m
	}
	for _, k := range msg.value(MSG_OPTION) {
		switch k {
		case MSG_CMDS, MSG_FIELDS, MSG_SESSID, EVENT:
			continue
		}
		if strings.HasSuffix(k, ".cb") {
			continue
		}
		if kit.IndexOf(m.value(MSG_APPEND), k) > -1 {
			continue
		}
		unlock := m.lock.Lock()
		if v, ok := msg._data[k]; ok {
			m._data[k] = v
			unlock()
		} else {
			unlock()
			m.Set(MSG_OPTION, k)
		}
		m.Add(MSG_OPTION, kit.Simple(k, msg.value(k))...)
	}
	kit.For(msg.value(MSG_APPEND), func(k string) { m.Add(MSG_APPEND, kit.Simple(k, msg.value(k))...) })
	return m.Add(MSG_RESULT, msg.value(MSG_RESULT)...)
}
func (m *Message) Length() (max int) {
	if m.FieldsIsDetail() {
		if len(m.value(KEY)) > 0 {
			return 1
		}
		return 0
	}
	kit.For(m.value(MSG_APPEND), func(k string) { max = kit.Max(len(m.value(k)), max) })
	return max
}
func (m *Message) TablesLimit(count int, cb func(value Maps)) *Message {
	return m.Table(func(value Maps, index int) { kit.If(index < count, func() { cb(value) }) })
}
func (m *Message) Stats(arg ...string) (res []string) {
	stats := map[string]float64{}
	m.Table(func(value Maps) { kit.For(arg, func(k string) { stats[k] += kit.Float(value[k]) }) })
	kit.For(arg, func(k string) { res = append(res, k, kit.Format("%0.2f", stats[k])) })
	return
}
func (m *Message) TableStats(field ...string) map[string]int {
	stat := map[string]int{}
	m.Table(func(value Maps) { kit.For(field, func(k string) { stat[value[k]]++ }) })
	return stat
}
func (m *Message) TableAmount(cb func(Maps) float64) float64 {
	var amount float64
	m.Table(func(value Maps) { amount += cb(value) })
	return amount
}
func (m *Message) Table(cb Any) *Message {
	n := m.Length()
	if n == 0 {
		return m
	}
	cbs := func(index int, value Maps, head []string) {
		switch cb := cb.(type) {
		case func(value Maps, index int, head []string):
			cb(value, index, head)
		case func(value Maps, index, total int):
			cb(value, index, n)
		case func(value Maps, index int):
			cb(value, index)
		case func(value Maps):
			cb(value)
		default:
			m.ErrorNotImplement(cb)
		}
	}
	if m.FieldsIsDetail() {
		value := Maps{}
		kit.For(m.value(KEY), func(i int, k string) { value[k] = kit.Select("", m.value(VALUE), i) })
		cbs(0, value, m.value(KEY))
		return m
	}
	for i := 0; i < n; i++ {
		value := Maps{}
		kit.For(m.value(MSG_APPEND), func(k string) { value[k] = kit.Select("", m.value(k), i) })
		cbs(i, value, m.value(MSG_APPEND))
	}
	return m
}
func (m *Message) TableEcho() *Message {
	const (
		TABLE_ROW_SEP = "table.row_sep"
		TABLE_COL_SEP = "table.col_sep"
		TABLE_SPACE   = "table.space"
		TABLE_ALIGN   = "table.align"
	)
	rows := kit.Select(NL, m.Option(TABLE_ROW_SEP))
	cols := kit.Select(SP, m.Option(TABLE_COL_SEP))
	show := func(value []string) {
		for i, v := range value {
			if m.Echo(v); i < len(value)-1 {
				m.Echo(cols)
			}
		}
		m.Echo(rows)
	}
	space := kit.Select(SP, m.Option(TABLE_SPACE))
	align := kit.Select("left", m.Option(TABLE_ALIGN))
	_align := func(value string, width int) string {
		switch n := width - kit.Width(value, len(space)); align {
		case "left":
			return value + strings.Repeat(space, n)
		case "right":
			return strings.Repeat(space, n) + value
		case "center":
			return strings.Repeat(space, n/2) + value + strings.Repeat(space, n-n/2)
		default:
			return value + space
		}
	}
	length, width := 0, map[string]int{}
	for _, k := range m.value(MSG_APPEND) {
		kit.If(len(m.value(k)) > length, func() { length = len(m.value(k)) })
		width[k] = kit.Width(k, len(space))
		kit.For(m.value(k), func(v string) {
			kit.If(kit.Width(v, len(space)) > width[k], func() { width[k] = kit.Width(v, len(space)) })
		})
	}
	show(kit.Simple(m.value(MSG_APPEND), func(k string) string { return _align(k, width[k]) }))
	for i := 0; i < length; i++ {
		show(kit.Simple(m.value(MSG_APPEND), func(k string) string { return _align(kit.Select("", m.value(k), i), width[k]) }))
	}
	return m
}
func (m *Message) TableEchoWithStatus() *Message {
	m.TableEcho()
	list := []string{}
	kit.For(kit.UnMarshal(m.Option(MSG_STATUS)), func(index int, value Map) {
		kit.If(value[VALUE] != nil, func() { list = append(list, kit.Format("%s: %s", value[NAME], value[VALUE])) })
	})
	kit.If(len(list) > 0, func() { m.Echo(strings.Join(list, SP)).Echo(NL) })
	return m
}
func (m *Message) Sort(key string, arg ...Any) *Message {
	if m.FieldsIsDetail() {
		key := m.value(KEY)
		value := m.value(VALUE)
		for i := 0; i < len(key)-1; i++ {
			for j := i+1; j < len(key); j++ {
				if key[i] > key[j] {
					key[i], key[j] = key[j], key[i]
					value[i], value[j] = value[j], value[i]
				}
			}
		}
		return m
	}
	order := map[string]map[string]int{}
	keys, cmps := kit.Split(kit.Select("type,name,text", key)), kit.Simple()
	for i, k := range keys {
		cmp := ""
		if i < len(arg) {
			switch v := arg[i].(type) {
			case string:
				cmp = v
			case []string:
				list := map[string]int{}
				for i, v := range v {
					list[v] = i + 1
				}
				order[k] = list
			case map[string]int:
				order[k] = v
			case func(string) int:
				list := map[string]int{}
				kit.For(m.Appendv(k), func(k string) {
					if _, ok := list[k]; !ok {
						list[k] = v(k)
					}
				})
				order[k] = list
			}
		}
		if cmp == "" {
			cmp = INT
			for _, v := range m.value(k) {
				if _, e := strconv.Atoi(v); e != nil {
					cmp = STR
				}
			}
		}
		cmps = append(cmps, cmp)
	}
	list := []Maps{}
	m.Table(func(value Maps) { list = append(list, value) })
	gt := func(i, j int) bool {
		for s, k := range keys {
			if a, b := list[i][k], list[j][k]; a != b {
				if v, ok := order[k]; ok {
					kit.If(v[a] == 0, func() { v[a] = len(v) + 1 })
					kit.If(v[b] == 0, func() { v[b] = len(v) + 1 })
					if v[a] > v[b] {
						return true
					} else if v[a] < v[b] {
						return false
					} else {
						continue
					}
				}
				switch cmp := cmps[s]; cmp {
				case STR, STR_R:
					if a > b {
						return cmp == STR
					} else if a < b {
						return cmp == STR_R
					}
				case INT, INT_R:
					if kit.Int(a) > kit.Int(b) {
						return cmp == INT
					} else if kit.Int(a) < kit.Int(b) {
						return cmp == INT_R
					}
				}
			}
		}
		return false
	}
	for i := 0; i < len(list)-1; i++ {
		min := i
		for j := i + 1; j < len(list); j++ {
			kit.If(gt(min, j), func() { min = j })
		}
		for j := min; j > i; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
	kit.For(m.value(MSG_APPEND), func(k string) { m.delete(k) })
	for _, v := range list {
		kit.For(m.value(MSG_APPEND), func(k string) { m.Add(MSG_APPEND, k, v[k]) })
	}
	return m
}
func (m *Message) SortStr(key string) *Message  { return m.Sort(key, STR) }
func (m *Message) SortStrR(key string) *Message { return m.Sort(key, STR_R) }
func (m *Message) SortIntR(key string) *Message { return m.Sort(key, INT_R) }
func (m *Message) SortInt(key string) *Message  { return m.Sort(key, INT) }

func (m *Message) Detail(arg ...Any) string { return m.index(MSG_DETAIL, 0) }
func (m *Message) Detailv(arg ...Any) []string {
	kit.If(len(arg) > 0, func() { m.value(MSG_DETAIL, kit.Simple(arg...)...) })
	return m.value(MSG_DETAIL)
}
func (m *Message) Options(arg ...Any) *Message {
	for i := 0; i < len(arg); i += 2 {
		if key, ok := arg[i].(string); ok {
			kit.If(i+1 < len(arg), func() { m.Optionv(key, arg[i+1]) })
		} else {
			kit.For(arg[i], func(k, v string) { m.Optionv(k, v) })
			i--
		}
	}
	return m
}
func (m *Message) Option(key string, arg ...Any) string {
	return kit.Select("", kit.Simple(m.Optionv(key, arg...)), 0)
}
func (m *Message) Append(key string, arg ...Any) string {
	if key == "" {
		return m.Append(m.Appendv(MSG_APPEND)[0])
	}
	return kit.Select("", m.Appendv(key, arg...), 0)
}
func (m *Message) Appendv(key string, arg ...Any) []string {
	if m.FieldsIsDetail() {
		if key == KEY {
			return m.value(key)
		}
		for i, k := range m.value(KEY) {
			if k == key || k == kit.Keys(EXTRA, key) {
				kit.If(len(arg) > 0, func() { m.index(VALUE, i, kit.Format(arg[0])) })
				return []string{m.index(VALUE, i)}
			}
		}
		if len(arg) > 0 {
			m.Add(MSG_APPEND, KEY, key).Add(MSG_APPEND, VALUE, kit.Format(arg[0]))
			return []string{kit.Format(arg[0])}
		}
		return nil
	}
	if key == MSG_APPEND {
		kit.If(len(arg) > 0, func() { m.value(key, kit.Simple(arg)...) })
		return m.value(key)
	}
	kit.If(len(arg) > 0, func() { m.value(key, kit.Simple(arg...)...) })

	defer m.lock.RLock()()
	if v, ok := m._meta[key]; ok {
		return v
	}
	if v, ok := m._meta[kit.Keys(EXTRA, key)]; ok {
		return v
	}
	return nil
}
func (m *Message) Resultv(arg ...Any) []string {
	if len(arg) > 0 {
		defer m.lock.Lock()()
		m._meta[MSG_RESULT] = kit.Simple(arg...)
	} else {
		defer m.lock.RLock()()
	}
	return m._meta[MSG_RESULT]
}
func (m *Message) Result(arg ...Any) string { return strings.Join(m.Resultv(arg...), "") }
func (m *Message) Results(arg ...Any) string {
	return kit.Select("", strings.TrimSpace(m.Result(arg...)), !m.IsErr())
}
