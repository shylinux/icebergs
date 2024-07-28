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
func (m *Message) OptionDefault(arg ...Any) string {
	kit.For(arg, func(k string, v Any) {
		switch v := v.(type) {
		case string:
			kit.If(m.Option(k) == "" && v != "", func() { m.Option(k, v) })
		case func() string:
			kit.If(m.Option(k) == "", func() { m.Option(k, v()) })
		}
	})
	return m.Option(kit.Format(arg[0]))
}
func (m *Message) OptionSimple(key ...string) (res []string) {
	kit.If(len(key) == 0, func() {
		key = kit.Filters(kit.Split(kit.Select("type,name,text", m.Conf(m.ShortKey(), kit.Keym(FIELD)))), TIME, HASH)
	})
	kit.For(kit.Filters(kit.Split(kit.Join(key)), ""), func(k string) { res = append(res, k, m.Option(k)) })
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
func (m *Message) ParseLink(p string) *Message {
	u := kit.ParseURL(p)
	switch arg := strings.Split(strings.TrimPrefix(u.Path, PS), PS); arg[0] {
	case "s":
		m.Option(MSG_USERPOD, kit.Select("", arg, 1))
	}
	m.Option(MSG_USERWEB, p)
	return m
}
func (m *Message) MergePod(pod string, arg ...Any) string {
	ls := []string{}
	kit.If(kit.Keys(m.Option(MSG_USERPOD), pod), func(p string) { ls = append(ls, "s", p) })
	return m.MergeLink(PS+path.Join(ls...), arg...)
}
func (m *Message) MergePodCmd(pod, cmd string, arg ...Any) string {
	ls := []string{}
	kit.If(kit.Keys(m.Option(MSG_USERPOD), pod), func(p string) { ls = append(ls, "s", p) })
	ls = append(ls, "c", kit.Select(m.ShortKey(), cmd))
	return m.MergeLink(PS+path.Join(ls...), arg...)
}
func (m *Message) MergeLink(url string, arg ...Any) string {
	kit.If(m.Option(DEBUG) == TRUE, func() { arg = append([]Any{DEBUG, TRUE}, arg...) })
	return kit.MergeURL2(strings.Split(kit.Select(Info.Domain, m.Option(MSG_USERHOST), m.Option(MSG_USERWEB)), QS)[0], url, arg...)
}
func (m *Message) FieldsSetDetail() {
	m.OptionFields(FIELDS_DETAIL)
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
	kit.If(len(arg) == 1 && arg[0] == nil, func() { arg = arg[:0] })
	kit.For(arg, func(i int, v Any) { arg[i] = kit.LowerCapital(kit.Format(v)) })
	return m.Options(MSG_ACTION, kit.Format(arg))
}
func (m *Message) Status(arg ...Any) *Message {
	list := kit.List()
	kit.For(kit.Simple(arg), func(k, v string) { kit.If(k, func() { list = append(list, kit.Dict(NAME, k, VALUE, v)) }) })
	return m.Options(MSG_STATUS, kit.Format(list))
}
func (m *Message) StatusTime(arg ...Any) *Message {
	args := []string{}
	kit.If(m.Option(MSG_DEBUG) == TRUE, func() { args = append(args, kit.MDB_COST, m.FormatCost(), "msg", "") })
	kit.If(m.Option(MSG_DEBUG) == TRUE, func() { args = append(args, m.OptionSimple(LOG_TRACEID)...) })
	return m.Status(TIME, m.Time(), arg, args)
}
func (m *Message) StatusTimeCount(arg ...Any) *Message {
	return m.StatusTime(append([]Any{kit.MDB_COUNT, kit.Split(m.FormatSize())[0]}, arg...))
}
func (m *Message) StatusTimeCountStats(field ...string) *Message {
	return m.StatusTimeCount(m.TableStats(field...))
}
func (m *Message) StatusTimeCountTotal(arg ...Any) *Message {
	if len(arg) > 0 && arg[0] != nil {
		return m.StatusTimeCount(append([]Any{kit.MDB_TOTAL}, arg...))
	} else {
		return m
	}
}
func (m *Message) Process(cmd string, arg ...Any) *Message {
	if len(arg) == 0 {
		return m.Options(MSG_PROCESS, cmd)
	} else {
		return m.Options(MSG_PROCESS, cmd, PROCESS_ARG, kit.Simple(arg...))
	}
}
func (m *Message) ProcessCookie(arg ...Any)   { m.Process(PROCESS_COOKIE, arg...) }
func (m *Message) ProcessSession(arg ...Any)  { m.Process(PROCESS_SESSION, arg...) }
func (m *Message) ProcessLocation(arg ...Any) { m.Process(PROCESS_LOCATION, arg...) }
func (m *Message) ProcessReplace(url string, arg ...Any) {
	m.Process(PROCESS_REPLACE, m.MergeLink(url, arg...))
}
func (m *Message) ProcessHistory(arg ...Any) { m.Process(PROCESS_HISTORY, arg...) }
func (m *Message) ProcessConfirm(arg ...Any) { m.Process(PROCESS_CONFIRM, arg...) }
func (m *Message) ProcessRefresh(arg ...string) *Message {
	return m.Process(PROCESS_REFRESH).Options(PROCESS_ARG, int(kit.Duration(kit.Select("30ms", arg, 0))/time.Millisecond))
}
func (m *Message) ProcessRewrite(arg ...Any) { m.Process(PROCESS_REWRITE, arg...) }
func (m *Message) ProcessDisplay(arg ...Any) { m.Process(PROCESS_DISPLAY).Option(MSG_DISPLAY, arg...) }
func (m *Message) ProcessField(arg ...Any)   { m.Process(PROCESS_FIELD).Option(FIELD_PREFIX, arg...) }
func (m *Message) ProcessInner() *Message    { return m.Process(PROCESS_INNER) }
func (m *Message) ProcessAgain()             { m.Process(PROCESS_AGAIN) }
func (m *Message) ProcessHold(text ...Any)   { m.Process(PROCESS_HOLD, text...) }
func (m *Message) ProcessBack(arg ...Any)    { m.Process(PROCESS_BACK, arg...) }
func (m *Message) ProcessRich(arg ...Any)    { m.Process(PROCESS_RICH, arg...) }
func (m *Message) ProcessGrow(arg ...Any)    { m.Process(PROCESS_GROW, arg...) }
func (m *Message) ProcessOpen(url string)    { m.Process(PROCESS_OPEN, url) }
func (m *Message) ProcessClose() *Message    { return m.Process(PROCESS_CLOSE) }
func (m *Message) ProcessOpenAndRefresh(url string) *Message {
	return m.Process(PROCESS_OPEN, url, "refresh")
}
