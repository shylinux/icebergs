package ice

import (
	"time"

	kit "shylinux.com/x/toolkits"
)

type Option struct {
	Name  string
	Value Any
}

func OptionFields(arg ...string) Option { return Option{MSG_FIELDS, kit.Join(arg)} }
func (m *Message) OptionFields(arg ...string) string {
	if len(arg) > 0 {
		m.Option(MSG_FIELDS, kit.Join(arg))
	}
	return kit.Join(kit.Simple(m.Optionv(MSG_FIELDS)))
}
func (m *Message) OptionFromConfig(arg ...string) string {
	for _, key := range arg {
		m.Option(key, m.Config(key))
	}
	return m.Option(arg[0])
}
func (m *Message) OptionDefault(arg ...string) string {
	for i := 0; i < len(arg); i += 2 {
		if m.Option(arg[i]) == "" && arg[i+1] != "" {
			m.Option(arg[i], arg[i+1])
		}
	}
	return m.Option(arg[0])
}
func (m *Message) OptionSimple(key ...string) (res []string) {
	if len(key) == 0 {
		for _, k := range kit.Split(kit.Select("type,name,text", m.Config(FIELD))) {
			switch k {
			case TIME, HASH:
				continue
			}
			if k == "" || m.Option(k) == "" {
				continue
			}
			res = append(res, k, m.Option(k))
		}
		return
	}
	for _, k := range kit.Split(kit.Join(key)) {
		if k == "" || m.Option(k) == "" {
			continue
		}
		res = append(res, k, m.Option(k))
	}
	return
}
func (m *Message) OptionSplit(key ...string) (res []string) {
	for _, k := range kit.Split(kit.Join(key)) {
		res = append(res, m.Option(k))
	}
	return res
}
func (m *Message) OptionCB(key string, cb ...Any) Any {
	if len(cb) > 0 {
		return m.Optionv(kit.Keycb(kit.Select(m.CommandKey(), key)), cb...)
	}
	return m.Optionv(kit.Keycb(kit.Select(m.CommandKey(), key)))
}

func (m *Message) FieldsIsDetail() bool {
	if len(m.meta[MSG_APPEND]) == 2 && m.meta[MSG_APPEND][0] == KEY && m.meta[MSG_APPEND][1] == VALUE {
		return true
	}
	if m.OptionFields() == FIELDS_DETAIL {
		return true
	}
	return false
}
func (m *Message) Fields(length int, fields ...string) string {
	return m.Option(MSG_FIELDS, kit.Select(kit.Select(FIELDS_DETAIL, fields, length), m.Option(MSG_FIELDS)))
}
func (m *Message) Action(arg ...Any) *Message {
	for i, v := range arg {
		switch v.(type) {
		case string:
		default:
			arg[i] = kit.Format(v)
		}
	}
	m.Option(MSG_ACTION, kit.Format(arg))
	return m
}
func (m *Message) Status(arg ...Any) *Message {
	list, args := kit.List(), kit.Simple(arg)
	for i := 0; i < len(args)-1; i += 2 {
		switch args[i+1] {
		case "", "0":
			continue
		}
		list = append(list, kit.Dict(NAME, args[i], VALUE, args[i+1]))
	}
	m.Option(MSG_STATUS, kit.Format(list))
	return m
}
func (m *Message) StatusTime(arg ...Any) *Message {
	return m.Status(TIME, m.Time(), arg, kit.MDB_COST, m.FormatCost())
}
func (m *Message) StatusTimeCount(arg ...Any) *Message {
	return m.Status(TIME, m.Time(), kit.MDB_COUNT, kit.Split(m.FormatSize())[0], arg, kit.MDB_COST, m.FormatCost())
}
func (m *Message) StatusTimeCountTotal(arg ...Any) *Message {
	return m.Status(TIME, m.Time(), kit.MDB_COUNT, kit.Split(m.FormatSize())[0], kit.MDB_TOTAL, arg, kit.MDB_COST, m.FormatCost())
}

func (m *Message) Process(cmd string, arg ...Any) {
	m.Option(MSG_PROCESS, cmd)
	m.Option(PROCESS_ARG, arg...)
}
func (m *Message) ProcessLocation(arg ...Any) {
	m.Process(PROCESS_LOCATION, arg...)
}
func (m *Message) ProcessReplace(url string, arg ...Any) {
	m.Process(PROCESS_REPLACE, kit.MergeURL(url, arg...))
}
func (m *Message) ProcessHistory(arg ...Any) {
	m.Process(PROCESS_HISTORY, arg...)
}
func (m *Message) ProcessConfirm(arg ...Any) {
	m.Process(PROCESS_CONFIRM, arg...)
}
func (m *Message) ProcessRefresh(arg ...string) {
	m.Process(PROCESS_REFRESH)
	if d, e := time.ParseDuration(kit.Select("30ms", arg, 0)); e == nil {
		m.Option(PROCESS_ARG, int(d/time.Millisecond))
	}
}
func (m *Message) ProcessRewrite(arg ...Any) {
	m.Process(PROCESS_REWRITE, arg...)
}
func (m *Message) ProcessDisplay(arg ...Any) {
	m.Process(PROCESS_DISPLAY)
	m.Option(MSG_DISPLAY, arg...)
}

func (m *Message) ProcessField(arg ...Any) {
	m.Process(PROCESS_FIELD)
	m.Option(FIELD_PREFIX, arg...)
}
func (m *Message) ProcessInner()           { m.Process(PROCESS_INNER) }
func (m *Message) ProcessAgain()           { m.Process(PROCESS_AGAIN) }
func (m *Message) ProcessHold(text ...Any) { m.Process(PROCESS_HOLD, text...) }
func (m *Message) ProcessBack()            { m.Process(PROCESS_BACK) }
func (m *Message) ProcessRich(arg ...Any)  { m.Process(PROCESS_RICH, arg...) }
func (m *Message) ProcessGrow(arg ...Any)  { m.Process(PROCESS_GROW, arg...) }
func (m *Message) ProcessOpen(url string)  { m.Process(PROCESS_OPEN, url) }
