package ice

import (
	"path"
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
func (m *Message) OptionArgs(key ...string) string {
	res := []string{}
	kit.For(kit.Split(kit.Join(key)), func(k string) { kit.If(m.Option(k), func(v string) { res = append(res, k, kit.Format("%q", v)) }) })
	return strings.Join(res, SP)
}
func (m *Message) OptionCB(key string, cb ...Any) Any {
	kit.If(len(cb) > 0, func() { m.Optionv(kit.Keycb(kit.Select(m.CommandKey(), key)), cb...) })
	return m.Optionv(kit.Keycb(kit.Select(m.CommandKey(), key)))
}

func (m *Message) MergePod(pod string, arg ...Any) string {
	ls := []string{"chat"}
	kit.If(kit.Keys(m.Option(MSG_USERPOD), pod), func(p string) { ls = append(ls, POD, p) })
	kit.If(len(ls) == 1, func() { ls = ls[:0] })
	// ls := []string{}
	// kit.If(kit.Keys(m.Option(MSG_USERPOD), pod), func(p string) { ls = append(ls, "/s/", p) })
	kit.If(m.Option(DEBUG) == TRUE, func() { arg = append([]Any{DEBUG, TRUE}, arg...) })
	return kit.MergeURL2(strings.Split(kit.Select("http://localhost:9020", Info.Domain, m.Option(MSG_USERWEB)), QS)[0], path.Join(PS, path.Join(ls...)), arg...)
}
func (m *Message) MergePodCmd(pod, cmd string, arg ...Any) string {
	ls := []string{"chat"}
	kit.If(kit.Keys(m.Option(MSG_USERPOD), pod), func(p string) { ls = append(ls, "s", p) })
	if cmd == "" {
		p, ok := Info.Index[m.CommandKey()]
		cmd = kit.Select(m.PrefixKey(), m.CommandKey(), ok && p == m.target)
	}
	ls = append(ls, "c", cmd)
	kit.If(m.Option(DEBUG) == TRUE, func() { arg = append([]Any{DEBUG, TRUE}, arg...) })
	return kit.MergeURL2(strings.Split(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), QS)[0], PS+kit.Join(ls[1:], PS), arg...)
	// return kit.MergeURL2(strings.Split(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), QS)[0], PS+kit.Join(ls, PS), arg...)
}
func (m *Message) FieldsIsDetail() bool {
	ls := m.value(MSG_APPEND)
	return len(ls) == 2 && ls[0] == KEY && ls[1] == VALUE || m.OptionFields() == FIELDS_DETAIL
}
func (m *Message) Fields(length int, fields ...string) string {
	kit.If(length >= len(fields), func() { m.Option(MSG_FIELDS, FIELDS_DETAIL) })
	return m.OptionDefault(MSG_FIELDS, kit.Select(FIELDS_DETAIL, fields, length))
}
func (m *Message) Action(arg ...Any) *Message {
	kit.For(arg, func(i int, v Any) { arg[i] = kit.LowerCapital(kit.Format(v)) })
	return m.Options(MSG_ACTION, kit.Format(arg))
}
func (m *Message) Status(arg ...Any) *Message {
	list := kit.List()
	kit.For(kit.Simple(arg), func(k, v string) { list = append(list, kit.Dict(NAME, k, VALUE, v)) })
	return m.Options(MSG_STATUS, kit.Format(list))
}
func (m *Message) StatusTime(arg ...Any) *Message {
	args := []string{}
	kit.If(m.Option(MSG_DEBUG) == TRUE, func() { args = append(args, kit.MDB_COST, m.FormatCost()) })
	kit.If(m.Option(MSG_DEBUG) == TRUE, func() { args = append(args, "msg", "") })
	kit.If(m.Option(MSG_DEBUG) == TRUE, func() { args = append(args, m.OptionSimple(LOG_TRACEID)...) })
	return m.Status(TIME, m.Time(), arg, args)
}
func (m *Message) StatusTimeCount(arg ...Any) *Message {
	return m.StatusTime(append([]Any{kit.MDB_COUNT, kit.Split(m.FormatSize())[0]}, arg...))
}
func (m *Message) StatusTimeCountTotal(arg ...Any) *Message {
	return m.StatusTimeCount(append([]Any{kit.MDB_TOTAL}, arg...))
}

func (m *Message) Process(cmd string, arg ...Any) *Message {
	if len(arg) == 0 {
		return m.Options(MSG_PROCESS, cmd)
	} else {
		return m.Options(MSG_PROCESS, cmd, PROCESS_ARG, kit.Simple(arg...))
	}
}
func (m *Message) ProcessCookie(arg ...Any) {
	m.Process(PROCESS_COOKIE, arg...)
}
func (m *Message) ProcessSession(arg ...Any) {
	m.Process(PROCESS_SESSION, arg...)
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
func (m *Message) ProcessInner() *Message    { return m.Process(PROCESS_INNER) }
func (m *Message) ProcessAgain()             { m.Process(PROCESS_AGAIN) }
func (m *Message) ProcessHold(text ...Any)   { m.Process(PROCESS_HOLD, text...) }
func (m *Message) ProcessBack()              { m.Process(PROCESS_BACK) }
func (m *Message) ProcessRich(arg ...Any)    { m.Process(PROCESS_RICH, arg...) }
func (m *Message) ProcessGrow(arg ...Any)    { m.Process(PROCESS_GROW, arg...) }
func (m *Message) ProcessOpen(url string)    { kit.If(url, func() { m.Process(PROCESS_OPEN, url) }) }
func (m *Message) ProcessClose()             { m.Process(PROCESS_CLOSE) }
