package ice

import (
	"strconv"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) setDetail(key string, arg ...string) *Message {
	for i := 0; i < len(m.meta[KEY]); i++ {
		if m.meta[KEY][i] == key {
			if len(arg) > 0 {
				m.meta[VALUE][i] = arg[0]
				break
			}
			for ; i < len(m.meta[KEY])-1; i++ {
				m.meta[KEY][i] = m.meta[KEY][i+1]
				m.meta[VALUE][i] = m.meta[VALUE][i+1]
			}
			m.meta[KEY] = kit.Slice(m.meta[KEY], 0, -1)
			m.meta[VALUE] = kit.Slice(m.meta[VALUE], 0, -1)
			break
		}
	}
	return m
}
func (m *Message) Set(key string, arg ...string) *Message {
	switch key {
	case MSG_DETAIL, MSG_RESULT:
		delete(m.meta, key)
	case MSG_OPTION, MSG_APPEND:
		if m.FieldsIsDetail() {
			if len(arg) > 0 {
				m.setDetail(arg[0], arg[1:]...)
			} else {
				delete(m.meta, KEY)
				delete(m.meta, VALUE)
				delete(m.meta, MSG_APPEND)
			}
		} else if len(arg) > 0 {
			if delete(m.meta, arg[0]); len(arg) > 1 {
				m.meta[arg[0]] = arg[1:]
			}
		} else {
			for _, k := range m.meta[key] {
				delete(m.meta, k)
			}
			delete(m.meta, key)
		}
		return m
	default:
		if m.FieldsIsDetail() {
			return m.setDetail(key, arg...)
		}
		for _, k := range kit.Split(key) {
			delete(m.meta, k)
		}
	}
	if len(arg) == 0 {
		return m
	}
	return m.Add(key, arg...)
}
func (m *Message) Add(key string, arg ...string) *Message {
	if len(arg) == 0 {
		return m
	}
	switch key {
	case MSG_DETAIL, MSG_RESULT:
		m.meta[key] = append(m.meta[key], arg...)
	case MSG_OPTION, MSG_APPEND:
		if index := 0; key == MSG_APPEND {
			if m.meta[MSG_OPTION], index = kit.SliceRemove(m.meta[MSG_OPTION], arg[0]); index > -1 {
				delete(m.meta, arg[0])
			}
		}
		if m.meta[arg[0]] = append(m.meta[arg[0]], arg[1:]...); kit.IndexOf(m.meta[key], arg[0]) == -1 {
			m.meta[key] = append(m.meta[key], arg[0])
		}
	}
	return m
}
func (m *Message) Cut(fields ...string) *Message {
	m.meta[MSG_APPEND] = kit.Split(kit.Join(fields))
	return m
}
func (m *Message) CutTo(key, to string) *Message {
	return m.Cut(key).RenameAppend(key, to)
}
func (m *Message) Push(key string, value Any, arg ...Any) *Message {
	head := kit.Simple()
	if len(head) == 0 && len(arg) > 0 {
		head = kit.Simple(arg[0])
	}
	if len(head) == 0 {
		head = kit.Simple(m.meta[MSG_APPEND])
	}
	if len(head) == 0 && !m.FieldsIsDetail() {
		head = kit.Split(m.OptionFields())
	}
	switch value := value.(type) {
	case Map:
		if len(head) == 0 {
			head = kit.SortedKey(kit.KeyValue(nil, "", value))
		}
		var val Map
		if len(arg) > 1 {
			val, _ = arg[1].(Map)
		}
		for _, k := range head {
			var v Any
			switch k {
			case "_target":
				continue
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
				m.Add(MSG_APPEND, KEY, strings.TrimPrefix(k, EXTRA+PT))
				m.Add(MSG_APPEND, VALUE, v)
			default:
				m.Add(MSG_APPEND, k, v)
			}
		}
	case Maps:
		if len(head) == 0 {
			head = kit.SortedKey(value)
		}
		for _, k := range head {
			m.Push(k, value[k])
		}
	default:
		for _, v := range kit.Simple(value, arg) {
			if m.FieldsIsDetail() {
				m.Add(MSG_APPEND, KEY, key)
				m.Add(MSG_APPEND, VALUE, kit.Format(value))
			} else {
				m.Add(MSG_APPEND, key, v)
			}
		}
	}
	return m
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
		for _, k := range arg[1:] {
			m.Add(arg[0], kit.Simple(k, msg.meta[k])...)
		}
		return m
	}
	for _, k := range msg.meta[MSG_OPTION] {
		switch k {
		case MSG_CMDS, MSG_FIELDS, MSG_SESSID, "event":
			continue
		}
		if strings.HasSuffix(k, ".cb") {
			continue
		}
		if kit.IndexOf(m.meta[MSG_APPEND], k) > -1 {
			continue
		}
		if v, ok := msg.data[k]; ok {
			m.data[k] = v
		} else {
			m.Set(MSG_OPTION, k)
		}
		m.Add(MSG_OPTION, kit.Simple(k, msg.meta[k])...)
	}
	for _, k := range msg.meta[MSG_APPEND] {
		m.Add(MSG_APPEND, kit.Simple(k, msg.meta[k])...)
	}
	return m.Add(MSG_RESULT, msg.meta[MSG_RESULT]...)
}
func (m *Message) Length() (max int) {
	for _, k := range m.meta[MSG_APPEND] {
		if l := len(m.meta[k]); l > max {
			max = l
		}
	}
	return max
}
func (m *Message) Tables(cbs ...func(value Maps)) *Message {
	return m.Table(func(index int, value Maps, head []string) {
		for _, cb := range cbs {
			if cb != nil {
				cb(value)
			}
		}
	})
}
func (m *Message) Table(cbs ...func(index int, value Maps, head []string)) *Message {
	if len(cbs) > 0 && cbs[0] != nil {
		n := m.Length()
		if n == 0 {
			return m
		}
		if m.FieldsIsDetail() {
			value := Maps{}
			for i, k := range m.meta[KEY] {
				value[k] = kit.Select("", m.meta[VALUE], i)
			}
			for _, cb := range cbs {
				cb(0, value, m.meta[KEY])
			}
			return m
		}
		for i := 0; i < n; i++ {
			value := Maps{}
			for _, k := range m.meta[MSG_APPEND] {
				value[k] = kit.Select("", m.meta[k], i)
			}
			for _, cb := range cbs {
				cb(i, value, m.meta[MSG_APPEND])
			}
		}
		return m
	}
	const (
		TABLE_ROW_SEP = "table.row_sep"
		TABLE_COL_SEP = "table.col_sep"
		TABLE_COMPACT = "table.compact"
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
	compact := m.Option(TABLE_COMPACT) == TRUE
	space := kit.Select(SP, m.Option(TABLE_SPACE))
	align := kit.Select("left", m.Option(TABLE_ALIGN))
	_align := func(value string, width int) string {
		if compact {
			return value + space
		}
		n := width - kit.Width(value, len(space))
		switch align {
		case "left":
			return value + strings.Repeat(space, n)
		case "right":
			return strings.Repeat(space, n) + value
		case "center":
			return strings.Repeat(space, n/2) + value + strings.Repeat(space, n-n/2)
		}
		return value + space
	}
	length, width := 0, map[string]int{}
	for _, k := range m.meta[MSG_APPEND] {
		if len(m.meta[k]) > length {
			length = len(m.meta[k])
		}
		width[k] = kit.Width(k, len(space))
		for _, v := range m.meta[k] {
			if kit.Width(v, len(space)) > width[k] {
				width[k] = kit.Width(v, len(space))
			}
		}
	}
	show(kit.Simple(m.meta[MSG_APPEND], func(k string) string { return _align(k, width[k]) }))
	for i := 0; i < length; i++ {
		show(kit.Simple(m.meta[MSG_APPEND], func(k string) string { return _align(kit.Select("", m.meta[k], i), width[k]) }))
	}
	return m
}

const (
	INT = "int"
	STR = "str"
	// TIME = "time"

	TIME_R = "time_r"
	STR_R  = "str_r"
	INT_R  = "int_r"
)

func (m *Message) Sort(key string, arg ...string) *Message {
	if m.FieldsIsDetail() {
		return m
	}
	keys, cmps := kit.Split(key), kit.Simple()
	for i, k := range keys {
		cmp := kit.Select("", arg, i)
		if cmp == "" {
			cmp = INT
			for _, v := range m.meta[k] {
				if _, e := strconv.Atoi(v); e != nil {
					cmp = STR
				}
			}
		}
		cmps = append(cmps, cmp)
	}
	list := []Maps{}
	m.Tables(func(value Maps) { list = append(list, value) })
	gt := func(i, j int) bool {
		for s, k := range keys {
			a, b := list[i][k], list[j][k]
			if a == b {
				continue
			}
			switch cmp := cmps[s]; cmp {
			case INT, INT_R:
				if kit.Int(a) > kit.Int(b) {
					return cmp == INT
				}
				if kit.Int(a) < kit.Int(b) {
					return cmp == INT_R
				}
			case STR, STR_R:
				if a > b {
					return cmp == STR
				}
				if a < b {
					return cmp == STR_R
				}
			case TIME, TIME_R:
				if kit.Time(a) > kit.Time(b) {
					return cmp == TIME
				}
				if kit.Time(a) < kit.Time(b) {
					return cmp == TIME_R
				}
			}
		}
		return false
	}
	for i := 0; i < len(list)-1; i++ {
		min := i
		for j := i + 1; j < len(list); j++ {
			if gt(min, j) {
				min = j
			}
		}
		for j := min; j > i; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
	for _, k := range m.meta[MSG_APPEND] {
		delete(m.meta, k)
	}
	for _, v := range list {
		for _, k := range m.meta[MSG_APPEND] {
			m.Add(MSG_APPEND, k, v[k])
		}
	}
	return m
}
func (m *Message) SortInt(key string)   { m.Sort(key, INT) }
func (m *Message) SortStr(key string)   { m.Sort(key, STR) }
func (m *Message) SortTime(key string)  { m.Sort(key, TIME) }
func (m *Message) SortTimeR(key string) { m.Sort(key, TIME_R) }
func (m *Message) SortStrR(key string)  { m.Sort(key, STR_R) }
func (m *Message) SortIntR(key string)  { m.Sort(key, INT_R) }

func (m *Message) Detail(arg ...Any) string {
	return kit.Select("", m.meta[MSG_DETAIL], 0)
}
func (m *Message) Detailv(arg ...Any) []string {
	if len(arg) > 0 {
		m.meta[MSG_DETAIL] = kit.Simple(arg...)
	}
	return m.meta[MSG_DETAIL]
}
func (m *Message) Options(arg ...Any) Any {
	for i := 0; i < len(arg); i += 2 {
		switch val := arg[i].(type) {
		case Maps:
			for k, v := range val {
				m.Optionv(k, v)
			}
			i--
			continue
		case []string:
			for i := 0; i < len(val)-1; i += 2 {
				m.Optionv(val[i], val[i+1])
			}
			i--
			continue
		}
		if i+1 < len(arg) {
			m.Optionv(kit.Format(arg[i]), arg[i+1])
		}
	}
	return m.Optionv(kit.Format(arg[0]))
}
func (m *Message) Optionv(key string, arg ...Any) Any {
	if len(arg) > 0 {
		if kit.IndexOf(m.meta[MSG_OPTION], key) == -1 {
			m.meta[MSG_OPTION] = append(m.meta[MSG_OPTION], key)
		}
		switch delete(m.data, key); v := arg[0].(type) {
		case nil:
			delete(m.meta, key)
		case string:
			m.meta[key] = kit.Simple(arg...)
		case []string:
			m.meta[key] = v
		default:
			m.data[key] = v
		}
	}
	for msg := m; msg != nil; msg = msg.message {
		if v, ok := msg.data[key]; ok {
			return v
		}
		if v, ok := msg.meta[key]; ok {
			return v
		}
	}
	return nil
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
		for i, k := range m.meta[KEY] {
			if k == key || k == kit.Keys(EXTRA, key) {
				if len(arg) > 0 {
					m.meta[VALUE][i] = kit.Format(arg[0])
				}
				return []string{kit.Select("", m.meta[VALUE], i)}
			}
		}
		if len(arg) > 0 {
			m.Add(MSG_APPEND, KEY, key)
			m.Add(MSG_APPEND, VALUE, kit.Format(arg[0]))
			return []string{kit.Format(arg[0])}
		}
		return nil
	}
	if key == MSG_APPEND {
		if len(arg) > 0 {
			m.meta[MSG_APPEND] = kit.Simple(arg)
		}
		return m.meta[key]
	}
	if len(arg) > 0 {
		m.meta[key] = kit.Simple(arg...)
	}
	if v, ok := m.meta[key]; ok {
		return v
	}
	if v, ok := m.meta[kit.Keys(EXTRA, key)]; ok {
		return v
	}
	return nil
}
func (m *Message) Resultv(arg ...Any) []string {
	if len(arg) > 0 {
		m.meta[MSG_RESULT] = kit.Simple(arg...)
	}
	return m.meta[MSG_RESULT]
}
func (m *Message) Result(arg ...Any) string {
	if len(arg) > 0 {
		switch v := arg[0].(type) {
		case int:
			return kit.Select("", m.meta[MSG_RESULT], v)
		}
	}
	return strings.Join(m.Resultv(arg...), "")
}
func (m *Message) Results(arg ...Any) string {
	return kit.Select("", strings.TrimSpace(m.Result()), !m.IsErr())
}
