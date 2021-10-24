package ice

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"time"

	kit "shylinux.com/x/toolkits"
)

type Option struct {
	Name  string
	Value interface{}
}

func OptionHash(arg string) Option      { return Option{kit.MDB_HASH, arg} }
func OptionFields(arg ...string) Option { return Option{MSG_FIELDS, kit.Join(arg)} }

func (m *Message) OptionFields(arg ...string) string {
	if len(arg) > 0 {
		m.Option(MSG_FIELDS, kit.Join(arg))
	}
	return kit.Join(kit.Simple(m.Optionv(MSG_FIELDS)))
}
func (m *Message) OptionPage(arg ...string) {
	m.Option(CACHE_LIMIT, kit.Select("10", arg, 0))
	m.Option(CACHE_OFFEND, kit.Select("0", arg, 1))
	m.Option(CACHE_FILTER, kit.Select("", arg, 2))
}
func (m *Message) OptionLoad(file string) *Message {
	if f, e := os.Open(file); e == nil {
		defer f.Close()

		var data interface{}
		json.NewDecoder(f).Decode(&data)

		kit.Fetch(data, func(key string, value interface{}) { m.Option(key, kit.Simple(value)) })
	}
	return m
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
	for _, key := range kit.Split(kit.MDB_STYLE) {
		if m.Option(key) != "" {
			res = append(res, kit.Format(`s="%s"`, key, m.Option(key)))
		}
	}
	for _, key := range kit.Split("type,name,text") {
		if m.Option(key) != "" {
			res = append(res, kit.Format(`data-%s="%s"`, key, m.Option(key)))
		}
	}
	kit.Fetch(m.Optionv(kit.MDB_EXTRA), func(key string, value string) {
		if value != "" {
			res = append(res, kit.Format(`data-%s="%s"`, key, value))
		}
	})
	return kit.Join(res, SP)
}

func (m *Message) Fields(length int, fields ...string) string {
	return m.Option(MSG_FIELDS, kit.Select(kit.Select("detail", fields, length), m.Option(MSG_FIELDS)))
}
func (m *Message) Upload(dir string) {
	up := kit.Simple(m.Optionv(MSG_UPLOAD))
	if len(up) < 2 {
		msg := m.Cmd("cache", "upload")
		up = kit.Simple(msg.Append(kit.MDB_HASH), msg.Append(kit.MDB_NAME), msg.Append(kit.MDB_SIZE))
	}

	if p := path.Join(dir, up[1]); m.Option(MSG_USERPOD) == "" {
		m.Cmdy("cache", "watch", up[0], p) // 本机文件
	} else { // 下发文件
		m.Cmdy("spide", DEV, SAVE, p, "GET", kit.MergeURL2(m.Option(MSG_USERWEB), path.Join("/share/cache", up[0])))
	}
}
func (m *Message) Action(arg ...string) {
	m.Option(MSG_ACTION, kit.Format(arg))
}
func (m *Message) Status(arg ...interface{}) {
	list := kit.List()
	args := kit.Simple(arg)
	for i := 0; i < len(args)-1; i += 2 {
		list = append(list, kit.Dict(kit.MDB_NAME, args[i], kit.MDB_VALUE, args[i+1]))
	}
	m.Option(MSG_STATUS, kit.Format(list))
}
func (m *Message) StatusTime(arg ...interface{}) {
	m.Status(kit.MDB_TIME, m.Time(), arg, kit.MDB_COST, m.FormatCost())
}
func (m *Message) StatusTimeCount(arg ...interface{}) {
	m.Status(kit.MDB_TIME, m.Time(), kit.MDB_COUNT, m.FormatSize(), arg, kit.MDB_COST, m.FormatCost())
}
func (m *Message) StatusTimeCountTotal(arg ...interface{}) {
	m.Status(kit.MDB_TIME, m.Time(), kit.MDB_COUNT, m.FormatSize(), kit.MDB_TOTAL, arg, kit.MDB_COST, m.FormatCost())
}

func (m *Message) Toast(text string, arg ...interface{}) { // [title [duration [progress]]]
	if len(arg) > 1 {
		switch val := arg[1].(type) {
		case string:
			if value, err := time.ParseDuration(val); err == nil {
				arg[1] = int(value / time.Millisecond)
			}
		}
	}

	if m.Option(MSG_USERPOD) == "" {
		m.Cmd("space", m.Option(MSG_DAEMON), "toast", "", text, arg)
	} else {
		m.Option(MSG_TOAST, kit.Simple(text, arg))
	}
}
func (m *Message) GoToast(title string, cb func(toast func(string, int, int))) {
	m.Go(func() {
		cb(func(name string, count, total int) {
			m.Toast(
				kit.Format("%s %s/%s", name, strings.TrimSuffix(kit.FmtSize(int64(count)), "B"), strings.TrimSuffix(kit.FmtSize(int64(total)), "B")),
				kit.Format("%s %d%%", title, count*100/total),
				kit.Select("1000", "10000", count < total),
				count*100/total,
			)
		})
	})
}

func (m *Message) Process(action string, arg ...interface{}) {
	m.Option(MSG_PROCESS, action)
	m.Option("_arg", arg...)
}
func (m *Message) ProcessLocation(arg ...interface{}) {
	m.Process(PROCESS_LOCATION, arg...)
}
func (m *Message) ProcessRewrite(arg ...interface{}) {
	m.Process(PROCESS_REWRITE, arg...)
}
func (m *Message) ProcessRefresh(delay string) {
	if d, e := time.ParseDuration(delay); e == nil {
		m.Option("_delay", int(d/time.Millisecond))
	}
	m.Process(PROCESS_REFRESH)
}
func (m *Message) ProcessRefresh30ms()  { m.ProcessRefresh("30ms") }
func (m *Message) ProcessRefresh300ms() { m.ProcessRefresh("300ms") }
func (m *Message) ProcessRefresh3s()    { m.ProcessRefresh("3s") }
func (m *Message) ProcessDisplay(arg ...interface{}) {
	m.Process(PROCESS_DISPLAY)
	m.Option(MSG_DISPLAY, arg...)
}

func (m *Message) ProcessCommand(cmd string, val []string, arg ...string) {
	if len(arg) > 0 && arg[0] == RUN {
		m.Cmdy(cmd, arg[1:])
		return
	}

	m.Cmdy("command", cmd)
	m.ProcessField(cmd, RUN)
	m.Push(ARG, kit.Format(val))
}
func (m *Message) ProcessCommandOpt(arg ...string) {
	m.Push(OPT, kit.Format(m.OptionSimple(arg...)))
}
func (m *Message) ProcessField(arg ...interface{}) {
	m.Process(PROCESS_FIELD)
	m.Option(FIELD_PREFIX, arg...)
}
func (m *Message) ProcessInner()          { m.Process(PROCESS_INNER) }
func (m *Message) ProcessAgain()          { m.Process(PROCESS_AGAIN) }
func (m *Message) ProcessOpen(url string) { m.Process(PROCESS_OPEN, url) }
func (m *Message) ProcessHold()           { m.Process(PROCESS_HOLD) }
func (m *Message) ProcessBack()           { m.Process(PROCESS_BACK) }
