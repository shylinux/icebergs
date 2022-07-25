package ice

import (
	"encoding/json"
	"net/url"
	"os"
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

func (m *Message) OptionCB(key string, cb ...Any) Any {
	if len(cb) > 0 {
		return m.Optionv(kit.Keycb(key), cb...)
	}
	return m.Optionv(kit.Keycb(key))
}
func (m *Message) OptionFields(arg ...string) string {
	if len(arg) > 0 {
		m.Option(MSG_FIELDS, kit.Join(arg))
	}
	return kit.Join(kit.Simple(m.Optionv(MSG_FIELDS)))
}
func (m *Message) OptionPages(arg ...string) (page int, size int) {
	m.Option(CACHE_LIMIT, kit.Select("", arg, 0))
	m.Option(CACHE_OFFEND, kit.Select("", arg, 1))
	m.Option(CACHE_FILTER, kit.Select("", arg, 2))
	m.Option("limit", kit.Select(m.Option("limit"), arg, 0))
	m.Option("offend", kit.Select(m.Option("offend"), arg, 1))
	size = kit.Int(kit.Select("10", m.Option("limit")))
	page = kit.Int(m.Option("offend"))/size + 1
	return
}
func (m *Message) OptionPage(arg ...string) int {
	page, _ := m.OptionPages(arg...)
	return page
}
func (m *Message) OptionLoad(file string) *Message {
	if f, e := os.Open(file); e == nil {
		defer f.Close()

		var data Any
		m.Assert(json.NewDecoder(f).Decode(&data))

		kit.Fetch(data, func(key string, value Any) {
			m.Option(key, kit.Simple(value))
		})
	}
	return m
}
func (m *Message) OptionDefault(key, value string) string {
	if m.Option(key) == "" {
		m.Option(key, value)
	}
	return m.Option(key)
}
func (m *Message) OptionSplit(key ...string) (res []string) {
	for _, k := range kit.Split(kit.Join(key)) {
		res = append(res, m.Option(k))
	}
	return res
}
func (m *Message) OptionSimple(key ...string) (res []string) {
	for _, k := range kit.Split(kit.Join(key)) {
		if k == "" || m.Option(k) == "" {
			continue
		}
		res = append(res, k, m.Option(k))
	}
	return
}
func (m *Message) OptionTemplate() string {
	res := []string{`class="story"`}
	for _, key := range kit.Split(STYLE) {
		if m.Option(key) != "" {
			res = append(res, kit.Format(`s="%s"`, key, m.Option(key)))
		}
	}
	for _, key := range kit.Split("type,name,text") {
		if key == TEXT && m.Option(TYPE) == "spark" {
			continue
		}
		if m.Option(key) != "" {
			res = append(res, kit.Format(`data-%s="%s"`, key, m.Option(key)))
		}
	}
	kit.Fetch(m.Optionv("extra"), func(key string, value string) {
		if value != "" {
			res = append(res, kit.Format(`data-%s="%s"`, key, value))
		}
	})
	return kit.Join(res, SP)
}

func (m *Message) FieldsIsDetail() bool {
	if len(m.meta[MSG_APPEND]) == 2 && m.meta[MSG_APPEND][0] == KEY && m.meta[MSG_APPEND][1] == VALUE {
		return true
	}
	if m.OptionFields() == CACHE_DETAIL {
		return true
	}
	return false
}
func (m *Message) Fields(length int, fields ...string) string {
	return m.Option(MSG_FIELDS, kit.Select(kit.Select(CACHE_DETAIL, fields, length), m.Option(MSG_FIELDS)))
}
func (m *Message) Upload(dir string) {
	up := kit.Simple(m.Optionv(MSG_UPLOAD))
	if len(up) < 2 {
		msg := m.Cmd(CACHE, "upload")
		up = kit.Simple(msg.Append(HASH), msg.Append(NAME), msg.Append("size"))
	}

	if p := path.Join(dir, up[1]); m.Option(MSG_USERPOD) == "" {
		m.Cmdy(CACHE, "watch", up[0], p) // 本机文件
	} else { // 下发文件
		m.Cmdy(SPIDE, DEV, SAVE, p, "GET", m.MergeURL2(path.Join("/share/cache", up[0])))
	}
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
func (m *Message) Status(arg ...Any) {
	list := kit.List()
	args := kit.Simple(arg)
	for i := 0; i < len(args)-1; i += 2 {
		list = append(list, kit.Dict(NAME, args[i], VALUE, args[i+1]))
	}
	m.Option(MSG_STATUS, kit.Format(list))
}
func (m *Message) StatusTime(arg ...Any) {
	m.Status(TIME, m.Time(), arg, kit.MDB_COST, m.FormatCost())
}
func (m *Message) StatusTimeCount(arg ...Any) {
	m.Status(TIME, m.Time(), kit.MDB_COUNT, kit.Split(m.FormatSize())[0], arg, kit.MDB_COST, m.FormatCost())
}
func (m *Message) StatusTimeCountTotal(arg ...Any) {
	m.Status(TIME, m.Time(), kit.MDB_COUNT, kit.Split(m.FormatSize())[0], kit.MDB_TOTAL, arg, kit.MDB_COST, m.FormatCost())
}

func (m *Message) ToastProcess(arg ...Any) func() {
	if len(arg) == 0 {
		arg = kit.List("", "-1")
	}
	if len(arg) == 1 {
		arg = append(arg, "-1")
	}
	m.Toast(PROCESS, arg...)
	return func() { m.Toast(SUCCESS) }
}
func (m *Message) ToastRestart(arg ...Any) { m.Toast(RESTART, arg...) }
func (m *Message) ToastFailure(arg ...Any) { m.Toast(FAILURE, arg...) }
func (m *Message) ToastSuccess(arg ...Any) { m.Toast(SUCCESS, arg...) }
func (m *Message) Toast(text string, arg ...Any) { // [title [duration [progress]]]
	if len(arg) > 1 {
		switch val := arg[1].(type) {
		case string:
			if value, err := time.ParseDuration(val); err == nil {
				arg[1] = int(value / time.Millisecond)
			}
		}
	}

	m.PushNoticeToast("", text, arg)
}
func (m *Message) PushNotice(arg ...Any) {
	m.Optionv(MSG_OPTS, m.meta[MSG_OPTION])
	if m.Option(MSG_USERPOD) == "" {
		m.Cmd(SPACE, m.Option(MSG_DAEMON), arg)
	} else {
		opts := kit.Dict(POD, m.Option(MSG_DAEMON), "cmds", kit.Simple(arg...))
		for _, k := range m.meta[MSG_OPTS] {
			opts[k] = m.Option(k)
		}
		m.Cmd("web.spide", OPS, m.MergeURL2("/share/toast/"), kit.Format(opts))
	}
}
func (m *Message) PushNoticeGrow(arg ...Any) {
	m.PushNotice(kit.List("grow", arg)...)
}
func (m *Message) PushNoticeToast(arg ...Any) {
	m.PushNotice(kit.List("toast", arg)...)
}
func (m *Message) PushRefresh(arg ...Any) {
	m.PushNotice(kit.List("refresh")...)
}
func (m *Message) Toast3s(text string, arg ...Any) {
	m.Toast(text, kit.List(kit.Select("", arg, 0), kit.Select("3s", arg, 1))...)
}
func (m *Message) Toast30s(text string, arg ...Any) {
	m.Toast(text, kit.List(kit.Select("", arg, 0), kit.Select("30s", arg, 1))...)
}
func (m *Message) GoToast(title string, cb func(toast func(string, int, int))) {
	m.Go(func() {
		cb(func(name string, count, total int) {
			m.Toast(
				kit.Format("%s %s/%s", name, strings.TrimSuffix(kit.FmtSize(int64(count)), "B"), strings.TrimSuffix(kit.FmtSize(int64(total)), "B")),
				kit.Format("%s %d%%", title, count*100/total),
				kit.Select("3000", "30000", count < total),
				count*100/total,
			)
		})
	})
}

func (m *Message) Process(action string, arg ...Any) {
	m.Option(MSG_PROCESS, action)
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
func (m *Message) ProcessRefresh(delay string) {
	if d, e := time.ParseDuration(delay); e == nil {
		m.Option("_delay", int(d/time.Millisecond))
	}
	m.Process(PROCESS_REFRESH)
}
func (m *Message) ProcessRefresh3ms()   { m.ProcessRefresh("3ms") }
func (m *Message) ProcessRefresh30ms()  { m.ProcessRefresh("30ms") }
func (m *Message) ProcessRefresh300ms() { m.ProcessRefresh("300ms") }
func (m *Message) ProcessRefresh3s()    { m.ProcessRefresh("3s") }
func (m *Message) ProcessRewrite(arg ...Any) {
	m.Process(PROCESS_REWRITE, arg...)
}
func (m *Message) ProcessDisplay(arg ...Any) {
	m.Process(PROCESS_DISPLAY)
	m.Option(MSG_DISPLAY, arg...)
}

func (m *Message) ProcessCommand(cmd string, val []string, arg ...string) {
	if len(arg) > 0 && arg[0] == RUN {
		m.Cmdy(cmd, arg[1:])
		return
	}

	m.Cmdy(COMMAND, cmd)
	m.ProcessField(cmd, RUN)
	m.Push(ARG, kit.Format(val))
}
func (m *Message) ProcessCommandOpt(arg []string, args ...string) {
	if len(arg) > 0 && arg[0] == RUN {
		return
	}
	m.Push("opt", kit.Format(m.OptionSimple(args...)))
}
func (m *Message) ProcessField(arg ...Any) {
	m.Process(PROCESS_FIELD)
	m.Option(FIELD_PREFIX, arg...)
}
func (m *Message) ProcessInner()          { m.Process(PROCESS_INNER) }
func (m *Message) ProcessAgain()          { m.Process(PROCESS_AGAIN) }
func (m *Message) ProcessHold()           { m.Process(PROCESS_HOLD) }
func (m *Message) ProcessBack()           { m.Process(PROCESS_BACK) }
func (m *Message) ProcessRich(arg ...Any) { m.Process(PROCESS_RICH, arg...) }
func (m *Message) ProcessGrow(arg ...Any) { m.Process(PROCESS_GROW, arg...) }
func (m *Message) ProcessOpen(url string) { m.Process(PROCESS_OPEN, url) }

func (m *Message) OptionUserWeb() *url.URL {
	return kit.ParseURL(m.Option(MSG_USERWEB))
}
func (m *Message) MergeURL2(url string, arg ...Any) string {
	return kit.MergeURL2(m.Option(MSG_USERWEB), url, arg...)
}
func (m *Message) MergeLink(url string, arg ...Any) string {
	return strings.Split(m.MergeURL2(url, arg...), "?")[0]
}
func (m *Message) MergePod(pod string, arg ...Any) string {
	return kit.MergePOD(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), pod, arg...)
}
func (m *Message) MergeCmd(cmd string, arg ...Any) string {
	if cmd == "" {
		cmd = m.PrefixKey()
	}
	if m.Option(MSG_USERPOD) == "" {
		return kit.MergeURL2(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), path.Join("/chat", "cmd", cmd))
	}
	return kit.MergeURL2(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), path.Join("/chat", "pod", m.Option(MSG_USERPOD), "cmd", cmd), arg...)
}
func (m *Message) MergeWebsite(web string, arg ...Any) string {
	if m.Option(MSG_USERPOD) == "" {
		return kit.MergeURL2(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), path.Join("/chat", "website", web))
	}
	return kit.MergeURL2(kit.Select(Info.Domain, m.Option(MSG_USERWEB)), path.Join("/chat", "pod", m.Option(MSG_USERPOD), "website", web), arg...)
}
