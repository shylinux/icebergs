package ice

import (
	"strings"
	"time"

	kit "shylinux.com/x/toolkits"
)

type Option struct {
	Name  string
	Value Any
}

func OptionFields(arg ...string) Option { return Option{MSG_FIELDS, kit.Join(arg)} }
func OptionSilent(arg ...string) Option { return Option{LOG_DISABLE, TRUE} }
func (m *Message) OptionFields(arg ...string) string {
	kit.If(len(arg) > 0, func() { m.Option(MSG_FIELDS, kit.Join(arg)) })
	return kit.Join(kit.Simple(m.Optionv(MSG_FIELDS)))
}
func (m *Message) OptionDefault(arg ...string) string {
	kit.For(arg, func(k, v string) { kit.If(m.Option(k) == "" && v != "", func() { m.Option(k, v) }) })
	return m.Option(arg[0])
}
func (m *Message) OptionSimple(key ...string) (res []string) {
	kit.If(len(key) == 0, func() {
		key = kit.Filters(kit.Split(kit.Select("type,name,text", m.Conf(m.PrefixKey(), kit.Keym(FIELD)))), TIME, HASH)
	})
	kit.For(kit.Filters(kit.Split(kit.Join(key)), ""), func(k string) { kit.If(m.Option(k), func(v string) { res = append(res, k, v) }) })
	return
}
func (m *Message) OptionSplit(key ...string) (res []string) {
	kit.For(kit.Split(kit.Join(key)), func(k string) { res = append(res, m.Option(k)) })
	return res
}
func (m *Message) OptionCB(key string, cb ...Any) Any {
	kit.If(len(cb) > 0, func() { m.Optionv(kit.Keycb(kit.Select(m.CommandKey(), key)), cb...) })
	return m.Optionv(kit.Keycb(kit.Select(m.CommandKey(), key)))
}

func (m *Message) MergePod(pod string, arg ...Any) string {
	ls := []string{"chat"}
	kit.If(kit.Keys(m.Option(MSG_USERPOD), pod), func(p string) { ls = append(ls, POD, p) })
	return kit.MergeURL2(strings.Split(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), QS)[0], PS+kit.Join(ls, PS), arg...)
}
func (m *Message) MergePodCmd(pod, cmd string, arg ...Any) string {
	ls := []string{"chat"}
	kit.If(kit.Keys(m.Option(MSG_USERPOD), pod), func(p string) { ls = append(ls, POD, p) })
	if cmd == "" {
		if _, ok := Info.Index[m.CommandKey()]; ok {
			cmd = m.CommandKey()
		} else {
			cmd = m.PrefixKey()
		}
	}
	ls = append(ls, CMD, cmd)
	return kit.MergeURL2(strings.Split(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), QS)[0], PS+kit.Join(ls, PS), arg...)
}
func (m *Message) FieldsIsDetail() bool {
	return len(m.meta[MSG_APPEND]) == 2 && m.meta[MSG_APPEND][0] == KEY && m.meta[MSG_APPEND][1] == VALUE || m.OptionFields() == FIELDS_DETAIL
}
func (m *Message) Fields(length int, fields ...string) string {
	return m.OptionDefault(MSG_FIELDS, kit.Select(FIELDS_DETAIL, fields, length))
}
func (m *Message) Action(arg ...Any) *Message {
	kit.For(arg, func(i int, v Any) { arg[i] = kit.Format(v) })
	return m.Options(MSG_ACTION, kit.Format(arg))
}
func (m *Message) Status(arg ...Any) *Message {
	list := kit.List()
	kit.For(kit.Simple(arg), func(k, v string) { list = append(list, kit.Dict(NAME, k, VALUE, v)) })
	return m.Options(MSG_STATUS, kit.Format(list))
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

func (m *Message) Process(cmd string, arg ...Any) *Message {
	return m.Options(MSG_PROCESS, cmd, PROCESS_ARG, kit.Simple(arg...))
}
func (m *Message) ProcessLocation(arg ...Any) { m.Process(PROCESS_LOCATION, arg...) }
func (m *Message) ProcessReplace(url string, arg ...Any) {
	m.Process(PROCESS_REPLACE, kit.MergeURL(url, arg...))
}
func (m *Message) ProcessHistory(arg ...Any) { m.Process(PROCESS_HISTORY, arg...) }
func (m *Message) ProcessConfirm(arg ...Any) { m.Process(PROCESS_CONFIRM, arg...) }
func (m *Message) ProcessRefresh(arg ...string) {
	m.Process(PROCESS_REFRESH).Option(PROCESS_ARG, int(kit.Duration(kit.Select("30ms", arg, 0))/time.Millisecond))
}
func (m *Message) ProcessRewrite(arg ...Any) { m.Process(PROCESS_REWRITE, arg...) }
func (m *Message) ProcessDisplay(arg ...Any) { m.Process(PROCESS_DISPLAY).Option(MSG_DISPLAY, arg...) }
func (m *Message) ProcessField(arg ...Any)   { m.Process(PROCESS_FIELD).Option(FIELD_PREFIX, arg...) }
func (m *Message) ProcessInner()             { m.Process(PROCESS_INNER) }
func (m *Message) ProcessAgain()             { m.Process(PROCESS_AGAIN) }
func (m *Message) ProcessHold(text ...Any)   { m.Process(PROCESS_HOLD, text...) }
func (m *Message) ProcessBack()              { m.Process(PROCESS_BACK) }
func (m *Message) ProcessRich(arg ...Any)    { m.Process(PROCESS_RICH, arg...) }
func (m *Message) ProcessGrow(arg ...Any)    { m.Process(PROCESS_GROW, arg...) }
func (m *Message) ProcessOpen(url string)    { m.Process(PROCESS_OPEN, url) }
