package ice

import (
	"strings"

	kit "shylinux.com/x/toolkits"
)

func (m *Message) Split(str string, arg ...string) *Message {
	m.Set(MSG_APPEND).Set(MSG_RESULT)
	field := kit.Select("", arg, 0)
	sp := kit.Select(SP, arg, 1)
	nl := kit.Select(NL, arg, 2)
	fields, indexs := kit.Split(field, sp, sp, sp), []int{}
	for i, l := range kit.Split(str, nl, nl, nl) {
		if strings.HasPrefix(l, "Binary") {
			continue
		}
		if strings.TrimSpace(l) == "" {
			continue
		}
		if i == 0 && (field == "" || field == INDEX) {
			if fields = kit.Split(l, sp, sp); field == INDEX {
				if strings.HasPrefix(l, SP) || strings.HasPrefix(l, TB) {
					indexs = append(indexs, 0)
					for _, v := range fields {
						indexs = append(indexs, strings.Index(l, v)+len(v))
					}
					indexs = indexs[0 : len(indexs)-1]
				} else {
					for _, v := range fields {
						indexs = append(indexs, strings.Index(l, v))
					}
				}
			}
			continue
		}
		if len(indexs) > 0 {
			for i, v := range indexs {
				if v >= len(l) {
					m.Push(strings.TrimSpace(kit.Select(SP, fields, i)), "")
					continue
				}
				if i == len(indexs)-1 {
					m.Push(strings.TrimSpace(kit.Select(SP, fields, i)), strings.TrimSpace(l[v:]))
				} else {
					m.Push(strings.TrimSpace(kit.Select(SP, fields, i)), strings.TrimSpace(l[v:indexs[i+1]]))
				}
			}
			continue
		}
		ls := kit.Split(l, sp, sp)
		for i, v := range ls {
			if i == len(fields)-1 {
				m.Push(kit.Select(SP, fields, i), strings.Join(ls[i:], sp))
				break
			}
			m.Push(kit.Select(SP, fields, i), v)
		}
	}
	return m
}
func (m *Message) SplitIndex(str string, arg ...string) *Message {
	return m.Split(str, kit.Simple(INDEX, arg)...)
}
func (m *Message) SetAppend(arg ...string) *Message {
	kit.If(len(arg) == 0, func() { m.OptionFields("") })
	return m.Set(MSG_APPEND, arg...)
}
func (m *Message) SetResult(arg ...string) *Message { return m.Set(MSG_RESULT, arg...) }
func (m *Message) PushRecord(value Any, arg ...string) *Message {
	return m.Push("", value, kit.Split(kit.Join(arg)))
}
func (m *Message) PushDetail(value Any, arg ...string) *Message {
	switch v := value.(type) {
	case string:
		value = kit.UnMarshal(v)
	}
	return m.Push(FIELDS_DETAIL, value, kit.Split(kit.Join(arg)))
}
func (m *Message) RenameOption(from, to string) *Message {
	return m.Options(to, m.Option(from), from, "")
}
func (m *Message) RenameAppend(arg ...string) *Message {
	kit.For(arg, func(from, to string) {
		if from == to {
			return
		}
		kit.For(m.meta[MSG_APPEND], func(i int, k string) {
			if k == from {
				m.meta[MSG_APPEND][i], m.meta[to] = to, m.meta[from]
				delete(m.meta, from)
			}
		})
	})
	return m
}
func (m *Message) ToLowerAppend(arg ...string) *Message {
	kit.For(m.meta[MSG_APPEND], func(k string) { m.RenameAppend(k, strings.ToLower(k)) })
	return m
}
func (m *Message) AppendSimple(key ...string) (res []string) {
	if len(key) == 0 {
		if m.FieldsIsDetail() {
			key = append(key, m.meta[KEY]...)
		} else {
			key = append(key, m.Appendv(MSG_APPEND)...)
		}
	}
	kit.For(kit.Split(kit.Join(key)), func(k string) { res = append(res, k, m.Append(k)) })
	return
}
func (m *Message) AppendTrans(cb func(value string, key string, index int) string) *Message {
	if m.FieldsIsDetail() {
		for i, v := range m.meta[VALUE] {
			k := m.meta[KEY][i]
			m.meta[VALUE][i] = cb(v, k, 0)
		}
		return m
	}
	for _, k := range m.meta[MSG_APPEND] {
		for i, v := range m.meta[k] {
			m.meta[k][i] = cb(v, k, i)
		}
	}
	return m
}
func (m *Message) CmdAppend(arg ...Any) string {
	args := kit.Simple(arg...)
	field := kit.Slice(args, -1)[0]
	return m._command(kit.Slice(args, 0, -1), OptionFields(field)).Append(field)
}
func (m *Message) CmdMap(arg ...string) map[string]map[string]string {
	field, list := kit.Slice(arg, -1)[0], map[string]map[string]string{}
	m._command(kit.Slice(arg, 0, -1)).Table(func(value Maps) { list[value[field]] = value })
	return list
}
