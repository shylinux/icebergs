package ice

import (
	kit "github.com/shylinux/toolkits"

	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func (m *Message) Add(key string, arg ...string) *Message {
	switch key {
	case MSG_DETAIL, MSG_RESULT:
		m.meta[key] = append(m.meta[key], arg...)

	case MSG_OPTION, MSG_APPEND:
		if len(arg) > 0 {
			if kit.IndexOf(m.meta[key], arg[0]) == -1 {
				m.meta[key] = append(m.meta[key], arg[0])
			}
			m.meta[arg[0]] = append(m.meta[arg[0]], arg[1:]...)
		}
	}
	return m
}
func (m *Message) Set(key string, arg ...string) *Message {
	switch key {
	case MSG_DETAIL, MSG_RESULT:
		delete(m.meta, key)
	case MSG_OPTION, MSG_APPEND:
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
		delete(m.meta, key)
		for _, k := range arg {
			delete(m.meta, k)
		}
	}
	return m.Add(key, arg...)
}
func (m *Message) Push(key string, value interface{}, arg ...interface{}) *Message {
	switch value := value.(type) {
	case map[string]string:
		for k, v := range value {
			m.Push(k, v)
		}
		return m
	case map[string]interface{}:
		if key == "detail" {
			// 格式转换
			value = kit.KeyValue(map[string]interface{}{}, "", value)
		}

		// 键值排序
		list := []string{}
		if len(arg) > 0 {
			list = kit.Simple(arg[0])
		} else {
			for k := range value {
				list = append(list, k)
			}
			sort.Strings(list)
		}

		var val map[string]interface{}
		if len(arg) > 1 {
			val, _ = arg[1].(map[string]interface{})
		}

		for _, k := range list {
			// 查找数据
			var v interface{}
			switch k {
			case kit.MDB_KEY, kit.MDB_ZONE:
				if key != "" {
					v = key
				}
				fallthrough
			default:
				if v = value[k]; v == nil {
					v = value["extra."+k]
				}
				if v == nil {
					if v = val[k]; v == nil {
						v = val["extra."+k]
					}
				}
			}

			// 追加数据
			switch v := kit.Format(v); key {
			case "detail":
				m.Add(MSG_APPEND, kit.MDB_KEY, k)
				m.Add(MSG_APPEND, kit.MDB_VALUE, v)
			default:
				m.Add(MSG_APPEND, k, v)
			}
		}
		return m
	}

	for _, v := range kit.Simple(value) {
		m.Add(MSG_APPEND, key, v)
	}
	return m
}
func (m *Message) Echo(str string, arg ...interface{}) *Message {
	if len(arg) > 0 {
		str = fmt.Sprintf(str, arg...)
	}
	m.meta[MSG_RESULT] = append(m.meta[MSG_RESULT], str)
	return m
}
func (m *Message) Copy(msg *Message, arg ...string) *Message {
	if m == msg {
		return m
	}
	if m == nil {
		return m
	}
	if len(arg) > 0 {
		// 精确复制
		for _, k := range arg[1:] {
			if kit.IndexOf(m.meta[arg[0]], k) == -1 {
				m.meta[arg[0]] = append(m.meta[arg[0]], k)
			}
			m.meta[k] = append(m.meta[k], msg.meta[k]...)
		}
		return m
	}

	// 复制选项
	for _, k := range msg.meta[MSG_OPTION] {
		if kit.IndexOf(m.meta[MSG_OPTION], k) == -1 {
			m.meta[MSG_OPTION] = append(m.meta[MSG_OPTION], k)
		}
		if _, ok := msg.meta[k]; ok {
			m.meta[k] = msg.meta[k]
		} else {
			m.data[k] = msg.data[k]
		}
	}

	// 复制数据
	for _, k := range msg.meta[MSG_APPEND] {
		if kit.IndexOf(m.meta[MSG_OPTION], k) > -1 && len(m.meta[k]) > 0 {
			m.meta[k] = m.meta[k][:0]
		}
		if kit.IndexOf(m.meta[MSG_APPEND], k) == -1 {
			m.meta[MSG_APPEND] = append(m.meta[MSG_APPEND], k)
		}
		m.meta[k] = append(m.meta[k], msg.meta[k]...)
	}

	// 复制文本
	m.meta[MSG_RESULT] = append(m.meta[MSG_RESULT], msg.meta[MSG_RESULT]...)
	return m
}
func (m *Message) Sort(key string, arg ...string) *Message {
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
	number := map[int]int{}
	table := []map[string]string{}
	m.Table(func(index int, line map[string]string, head []string) {
		table = append(table, line)
		switch cmp {
		case "int":
			number[index] = kit.Int(line[key])
		case "int_r":
			number[index] = -kit.Int(line[key])
		case "time":
			number[index] = kit.Time(line[key])
		case "time_r":
			number[index] = -kit.Time(line[key])
		}
	})

	// 排序数据
	for i := 0; i < len(table)-1; i++ {
		for j := i + 1; j < len(table); j++ {
			result := false
			switch cmp {
			case "", "str":
				if table[i][key] > table[j][key] {
					result = true
				}
			case "str_r":
				if table[i][key] < table[j][key] {
					result = true
				}
			default:
				if number[i] > number[j] {
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
		nrow := 0
		for _, k := range m.meta[MSG_APPEND] {
			if len(m.meta[k]) > nrow {
				nrow = len(m.meta[k])
			}
		}

		for i := 0; i < nrow; i++ {
			line := map[string]string{}
			for _, k := range m.meta[MSG_APPEND] {
				line[k] = kit.Select("", m.meta[k], i)
			}
			// 依次回调
			cbs[0](i, line, m.meta[MSG_APPEND])
		}
		return m
	}

	//计算列宽
	space := kit.Select(" ", m.Option("table.space"))
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
	rows := kit.Select("\n", m.Option("table.row_sep"))
	cols := kit.Select(" ", m.Option("table.col_sep"))
	compact := m.Option("table.compact") == "true"
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
	row := map[string]string{}
	wor := []string{}
	for _, k := range m.meta[MSG_APPEND] {
		row[k], wor = k, append(wor, k+strings.Repeat(space, width[k]-kit.Width(k, len(space))))
	}
	if !cb(row, wor, -1) {
		return m
	}

	// 输出数据
	for i := 0; i < depth; i++ {
		row := map[string]string{}
		wor := []string{}
		for _, k := range m.meta[MSG_APPEND] {
			data := ""
			if i < len(m.meta[k]) {
				data = m.meta[k][i]
			}

			row[k], wor = data, append(wor, data+strings.Repeat(space, width[k]-kit.Width(data, len(space))))
		}
		// 依次回调
		if !cb(row, wor, i) {
			break
		}
	}
	return m
}
func (m *Message) Render(cmd string, args ...interface{}) *Message {
	m.Log(LOG_EXPORT, "%s: %v", cmd, args)
	m.Optionv(MSG_OUTPUT, cmd)
	m.Optionv(MSG_ARGS, args)

	switch cmd {
	case RENDER_TEMPLATE:
		if len(args) == 1 {
			args = append(args, m)
		}
		if res, err := kit.Render(args[0].(string), args[1]); m.Assert(err) {
			m.Echo(string(res))
		}
	}
	return m
}
func (m *Message) Parse(meta string, key string, arg ...string) *Message {
	list := []string{}
	for _, line := range kit.Split(strings.Join(arg, " "), "\n") {
		ls := kit.Split(line)
		for i := 0; i < len(ls); i++ {
			if strings.HasPrefix(ls[i], "#") {
				ls = ls[:i]
				break
			}
		}
		list = append(list, ls...)
	}

	switch data := kit.Parse(nil, "", list...); meta {
	case MSG_OPTION:
		m.Option(key, data)
	case MSG_APPEND:
		m.Append(key, data)
	}
	return m
}
func (m *Message) Split(str string, field string, space string, enter string) *Message {
	indexs := []int{}
	fields := kit.Split(field, space, "{}")
	for i, l := range kit.Split(str, enter, "{}") {
		if i == 0 && (field == "" || field == "index") {
			// 表头行
			fields = kit.Split(l, space)
			if field == "index" {
				for _, v := range fields {
					indexs = append(indexs, strings.Index(l, v))
				}
			}
			continue
		}

		if len(indexs) > 0 {
			// 数据行
			for i, v := range indexs {
				if i == len(indexs)-1 {
					m.Push(kit.Select("some", fields, i), l[v:])
				} else {
					m.Push(kit.Select("some", fields, i), l[v:indexs[i+1]])
				}
			}
			continue
		}

		for i, v := range kit.Split(l, space) {
			m.Push(kit.Select("some", fields, i), v)
		}
	}
	return m
}
func (m *Message) CSV(text string, head ...string) *Message {
	bio := bytes.NewBufferString(text)
	r := csv.NewReader(bio)

	if len(head) == 0 {
		head, _ = r.Read()
	}
	for {
		line, e := r.Read()
		if e != nil {
			break
		}
		for i, k := range head {
			m.Push(k, kit.Select("", line, i))
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
		// 写数据
		if kit.IndexOf(m.meta[MSG_OPTION], key) == -1 {
			m.meta[MSG_OPTION] = append(m.meta[MSG_OPTION], key)
		}

		switch str := arg[0].(type) {
		case nil:
			delete(m.meta, key)
		case string:
			m.meta[key] = kit.Simple(arg)
		case []string:
			m.meta[key] = str
		default:
			m.data[key] = str
		}
	}

	for msg := m; msg != nil; msg = msg.message {
		if list, ok := msg.data[key]; ok {
			// 读数据
			return list
		}
		if list, ok := msg.meta[key]; ok {
			// 读选项
			return list
		}
	}
	return nil
}
func (m *Message) Options(key string, arg ...interface{}) bool {
	return kit.Select("", kit.Simple(m.Optionv(key, arg...)), 0) != ""
}
func (m *Message) Option(key string, arg ...interface{}) string {
	return kit.Select("", kit.Simple(m.Optionv(key, arg...)), 0)
}
func (m *Message) Append(key string, arg ...interface{}) string {
	return kit.Select("", m.Appendv(key, arg...), 0)
}
func (m *Message) Appendv(key string, arg ...interface{}) []string {
	if key == MSG_APPEND {
		m.meta[MSG_APPEND] = kit.Simple(arg)
		return m.meta[key]
	}
	if key == "_index" {
		max := 0
		for _, k := range m.meta[MSG_APPEND] {
			if len(m.meta[k]) > max {
				max = len(m.meta[k])
			}
		}
		index := []string{}
		for i := 0; i < max; i++ {
			index = append(index, kit.Format(i))
		}
		return index
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
