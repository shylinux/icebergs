package ice

import (
	"sort"
	"strconv"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) Set(key string, arg ...string) *Message {
	switch key {
	case MSG_DETAIL, MSG_RESULT:
		delete(m.meta, key)
	case MSG_OPTION, MSG_APPEND:
		if m.FieldsIsDetail() {
			if len(arg) > 0 {
				for i := 0; i < len(m.meta[KEY]); i++ {
					if m.meta[KEY][i] == arg[0] {
						m.meta[KEY][i] = ""
						m.meta[VALUE][i] = ""
					}
				}
				return m
			}
			delete(m.meta, KEY)
			delete(m.meta, VALUE)
			return m
		}
		if len(arg) > 0 {
			if delete(m.meta, arg[0]); len(arg) == 1 {
				return m
			}
		} else {
			for _, k := range m.meta[key] {
				delete(m.meta, k)
			}
			delete(m.meta, key)
			return m
		}
	default:
		for _, k := range kit.Split(key) {
			delete(m.meta, k)
		}
		if m.FieldsIsDetail() {
			for i := 0; i < len(m.meta[KEY]); i++ {
				if m.meta[KEY][i] == key {
					if len(arg) > 0 {
						m.meta[VALUE][i] = arg[0]
						return m
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
	}
	return m.Add(key, arg...)
}
func (m *Message) Add(key string, arg ...string) *Message {
	switch key {
	case MSG_DETAIL, MSG_RESULT:
		m.meta[key] = append(m.meta[key], arg...)

	case MSG_OPTION, MSG_APPEND:
		if len(arg) == 0 {
			break
		}
		if key == MSG_APPEND {
			if i := kit.IndexOf(m.meta[MSG_OPTION], arg[0]); i > -1 {
				m.meta[MSG_OPTION][i] = ""
				delete(m.meta, arg[0])
			}
			if kit.IndexOf(m.meta[key], arg[0]) == -1 {
				m.meta[arg[0]] = []string{}
			}
		}

		if kit.IndexOf(m.meta[key], arg[0]) == -1 {
			m.meta[key] = append(m.meta[key], arg[0])
		}
		m.meta[arg[0]] = append(m.meta[arg[0]], arg[1:]...)
	}
	return m
}
func (m *Message) Cut(fields ...string) *Message {
	m.meta[MSG_APPEND] = kit.Split(kit.Join(fields))
	return m
}
func (m *Message) Push(key string, value interface{}, arg ...interface{}) *Message {
	switch value := value.(type) {
	case map[string]interface{}:
		head := kit.Simple()
		if len(arg) > 0 {
			head = kit.Simple(arg[0])
		} else { // 键值排序
			for k := range kit.KeyValue(map[string]interface{}{}, "", value) {
				head = append(head, k)
			}
			sort.Strings(head)
		}

		var val map[string]interface{}
		if len(arg) > 1 {
			val, _ = arg[1].(map[string]interface{})
		}

		for _, k := range head {
			// 查找数据
			var v interface{}
			switch k {
			case KEY, HASH:
				if key != "" && key != "detail" {
					v = key
					break
				}
				fallthrough
			default:
				if v = value[k]; v == nil {
					if v = kit.Value(value, k); v == nil {
						if v = val[k]; v == nil {
							v = kit.Value(val, k)
						}
					}
				}
			}

			// 追加数据
			switch v := kit.Format(v); key {
			case "detail":
				m.Add(MSG_APPEND, KEY, k)
				m.Add(MSG_APPEND, VALUE, v)
			default:
				m.Add(MSG_APPEND, k, v)
			}
		}

	case map[string]string:
		head := kit.Simple()
		if len(arg) > 0 {
			head = kit.Simple(arg[0])
		} else { // 键值排序
			for k := range value {
				head = append(head, k)
			}
			sort.Strings(head)
		}

		for _, k := range head {
			m.Push(k, value[k])
		}

	default:
		if m.FieldsIsDetail() {
			if key != KEY || key != VALUE {
				m.Add(MSG_APPEND, KEY, key)
				m.Add(MSG_APPEND, VALUE, kit.Format(value))
				break
			}
		}
		for _, v := range kit.Simple(value) {
			m.Add(MSG_APPEND, key, v)
		}
	}
	return m
}
func (m *Message) Echo(str string, arg ...interface{}) *Message {
	if str == "" {
		return m
	}
	m.meta[MSG_RESULT] = append(m.meta[MSG_RESULT], kit.Format(str, arg...))
	return m
}
func (m *Message) Copy(msg *Message, arg ...string) *Message {
	if m == nil || m == msg {
		return m
	}
	if len(arg) > 0 {
		for _, k := range arg[1:] {
			m.Add(arg[0], kit.Simple(k, msg.meta[k])...)
		}
		return m
	}

	for _, k := range msg.meta[MSG_OPTION] {
		if v, ok := msg.data[k]; ok {
			m.data[k] = v
		} else {
			m.Set(MSG_OPTION, k)
			m.Add(MSG_OPTION, kit.Simple(k, msg.meta[k])...)
		}
	}
	for _, k := range msg.meta[MSG_APPEND] {
		m.Add(MSG_APPEND, kit.Simple(k, msg.meta[k])...)
	}
	m.meta[MSG_RESULT] = append(m.meta[MSG_RESULT], msg.meta[MSG_RESULT]...)
	return m
}
func (m *Message) Sort(key string, arg ...string) *Message {
	ls := kit.Split(key)
	key = ls[0]

	if m.FieldsIsDetail() && key != KEY {
		return m
	}
	// 排序方法
	cmp := "str"
	if len(arg) > 0 && arg[0] != "" {
		cmp = arg[0]
	} else {
		cmp = "int"
		for _, v := range m.meta[key] {
			if _, e := strconv.Atoi(v); e != nil {
				cmp = "str"
			}
		}
	}

	// 排序因子
	number := map[int]int64{}
	table := []map[string]string{}
	m.Table(func(index int, line map[string]string, head []string) {
		switch table = append(table, line); cmp {
		case "int":
			number[index] = kit.Int64(line[key])
		case "int_r":
			number[index] = -kit.Int64(line[key])
		case "time":
			number[index] = int64(kit.Time(line[key]))
		case "time_r":
			number[index] = -int64(kit.Time(line[key]))
		}
	})
	compare := func(i, j int, op string) bool {
		for k := range ls[1:] {
			if table[i][ls[k]] == table[j][ls[k]] {
				continue
			}

			if op == ">" && table[i][ls[k]] > table[j][ls[k]] {
				return true
			}
			if op == "<" && table[i][ls[k]] < table[j][ls[k]] {
				return true
			}
			return false
		}
		return false
	}

	// 排序数据
	for i := 0; i < len(table)-1; i++ {
		for j := i + 1; j < len(table); j++ {
			result := false
			switch cmp {
			case "", "str":
				if table[i][key] > table[j][key] {
					result = true
				} else if table[i][key] == table[j][key] && compare(i, j, ">") {
					result = true
				}
			case "str_r":
				if table[i][key] < table[j][key] {
					result = true
				} else if table[i][key] == table[j][key] && compare(i, j, "<") {
					result = true
				}
			default:
				if number[i] > number[j] {
					result = true
				} else if table[i][key] == table[j][key] && compare(i, j, ">") {
					result = true
				}
			}

			if result {
				table[i], table[j] = table[j], table[i]
				number[i], number[j] = number[j], number[i]
			}
		}
	}

	// 输出数据
	for _, k := range m.meta[MSG_APPEND] {
		delete(m.meta, k)
	}
	for _, v := range table {
		for _, k := range m.meta[MSG_APPEND] {
			m.Add(MSG_APPEND, k, v[k])
		}
	}
	return m
}
func (m *Message) Table(cbs ...func(index int, value map[string]string, head []string)) *Message {
	if len(cbs) > 0 && cbs[0] != nil {
		if m.FieldsIsDetail() {
			if m.Length() == 0 {
				return m
			}
			line := map[string]string{}
			for i, k := range m.meta[KEY] {
				line[k] = kit.Select("", m.meta[VALUE], i)
			}
			cbs[0](0, line, m.meta[KEY])
			return m
		}

		for i := 0; i < m.Length(); i++ {
			line := map[string]string{}
			for _, k := range m.meta[MSG_APPEND] {
				line[k] = kit.Select("", m.meta[k], i)
			}
			cbs[0](i, line, m.meta[MSG_APPEND])
		}
		return m
	}

	//计算列宽
	space := kit.Select(SP, m.Option("table.space"))
	depth, width := 0, map[string]int{}
	for _, k := range m.meta[MSG_APPEND] {
		if len(m.meta[k]) > depth {
			depth = len(m.meta[k])
		}
		width[k] = kit.Width(k, len(space))
		for _, v := range m.meta[k] {
			if kit.Width(v, len(space)) > width[k] {
				width[k] = kit.Width(v, len(space))
			}
		}
	}

	// 回调函数
	rows := kit.Select(NL, m.Option("table.row_sep"))
	cols := kit.Select(SP, m.Option("table.col_sep"))
	compact := m.Option("table.compact") == TRUE
	cb := func(value map[string]string, field []string, index int) bool {
		for i, v := range field {
			if k := m.meta[MSG_APPEND][i]; compact {
				v = value[k]
			}

			if m.Echo(v); i < len(field)-1 {
				m.Echo(cols)
			}
		}
		m.Echo(rows)
		return true
	}

	// 输出表头
	row, wor := map[string]string{}, []string{}
	for _, k := range m.meta[MSG_APPEND] {
		row[k], wor = k, append(wor, k+strings.Repeat(space, width[k]-kit.Width(k, len(space))))
	}
	if !cb(row, wor, -1) {
		return m
	}

	// 输出数据
	for i := 0; i < depth; i++ {
		row, wor := map[string]string{}, []string{}
		for _, k := range m.meta[MSG_APPEND] {
			data := ""
			if i < len(m.meta[k]) {
				data = m.meta[k][i]
			}

			row[k], wor = data, append(wor, data+strings.Repeat(space, width[k]-kit.Width(data, len(space))))
		}
		if !cb(row, wor, i) {
			break
		}
	}
	return m
}

func (m *Message) Detail(arg ...interface{}) string {
	return kit.Select("", m.meta[MSG_DETAIL], 0)
}
func (m *Message) Detailv(arg ...interface{}) []string {
	return m.meta[MSG_DETAIL]
}
func (m *Message) Optionv(key string, arg ...interface{}) interface{} {
	if len(arg) > 0 {
		if kit.IndexOf(m.meta[MSG_OPTION], key) == -1 { // 写数据
			m.meta[MSG_OPTION] = append(m.meta[MSG_OPTION], key)
		}

		switch delete(m.data, key); str := arg[0].(type) {
		case nil:
			delete(m.meta, key)
		case string:
			m.meta[key] = kit.Simple(arg...)
		case []string:
			m.meta[key] = str
		default:
			m.data[key] = str
		}
	}

	for msg := m; msg != nil; msg = msg.message {
		if list, ok := msg.data[key]; ok {
			return list // 读数据
		}
		if list, ok := msg.meta[key]; ok {
			return list // 读选项
		}
	}
	return nil
}
func (m *Message) Option(key string, arg ...interface{}) string {
	return kit.Select("", kit.Simple(m.Optionv(key, arg...)), 0)
}
func (m *Message) Append(key string, arg ...interface{}) string {
	return kit.Select("", m.Appendv(key, arg...), 0)
}
func (m *Message) Appendv(key string, arg ...interface{}) []string {
	if key == MSG_APPEND {
		if len(arg) > 0 {
			m.meta[MSG_APPEND] = kit.Simple(arg)
		}
		return m.meta[key]
	}

	if m.FieldsIsDetail() && key != KEY {
		for i, k := range m.meta[KEY] {
			if k == key {
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

	if len(arg) > 0 {
		m.meta[key] = kit.Simple(arg...)
	}
	return m.meta[key]
}
func (m *Message) Resultv(arg ...interface{}) []string {
	if len(arg) > 0 {
		m.meta[MSG_RESULT] = kit.Simple(arg...)
	}
	return m.meta[MSG_RESULT]
}
func (m *Message) Result(arg ...interface{}) string {
	if len(arg) > 0 {
		switch v := arg[0].(type) {
		case int:
			return kit.Select("", m.meta[MSG_RESULT], v)
		}
	}
	return strings.Join(m.Resultv(arg...), "")
}

func (m *Message) SortInt(key string)   { m.Sort(key, "int") }
func (m *Message) SortIntR(key string)  { m.Sort(key, "int_r") }
func (m *Message) SortStr(key string)   { m.Sort(key, "str") }
func (m *Message) SortStrR(key string)  { m.Sort(key, "str_r") }
func (m *Message) SortTime(key string)  { m.Sort(key, "time") }
func (m *Message) SortTimeR(key string) { m.Sort(key, "time_r") }
