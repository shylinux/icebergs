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

type Sort struct {
	Fields string
	Method string
}
type Option struct {
	Name  string
	Value interface{}
}

func OptionHash(str string) Option      { return Option{kit.MDB_HASH, str} }
func OptionFields(str ...string) Option { return Option{MSG_FIELDS, strings.Join(str, ",")} }
func (m *Message) OptionFields(str ...string) string {
	if len(str) > 0 {
		m.Option(MSG_FIELDS, strings.Join(str, ","))
	}
	return m.Option(MSG_FIELDS)
}
func (m *Message) OptionPage(arg ...string) {
	m.Option("cache.offend", kit.Select("0", arg, 1))
	m.Option("cache.limit", kit.Select("10", arg, 0))
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
func (m *Message) OptionSplit(fields ...string) (res []string) {
	for _, k := range strings.Split(strings.Join(fields, ","), ",") {
		res = append(res, m.Option(k))
	}
	return res
}
func (m *Message) OptionSimple(key ...string) (res []string) {
	for _, k := range strings.Split(strings.Join(key, ","), ",") {
		if k == "" || m.Option(k) == "" {
			continue
		}
		res = append(res, k, m.Option(k))
	}
	return
}
func (m *Message) OptionTemplate() string {
	res := []string{`class="story"`}
	for _, key := range kit.Split("style") {
		if m.Option(key) != "" {
			res = append(res, kit.Format(`s="%s"`, key, m.Option(key)))
		}
	}
	for _, key := range kit.Split("type,name,text") {
		if m.Option(key) != "" {
			res = append(res, kit.Format(`data-%s="%s"`, key, m.Option(key)))
		}
	}
	kit.Fetch(m.Optionv("extra"), func(key string, value string) {
		if value != "" {
			res = append(res, kit.Format(`data-%s="%s"`, key, value))
		}
	})
	return strings.Join(res, " ")
}

func (m *Message) Fields(length int, fields ...string) string {
	return m.Option(MSG_FIELDS, kit.Select(kit.Select("detail", fields, length), m.Option(MSG_FIELDS)))
}
func (m *Message) Upload(dir string) {
	up := kit.Simple(m.Optionv(MSG_UPLOAD))
	if p := path.Join(dir, up[1]); m.Option(MSG_USERPOD) == "" {
		// 本机文件
		m.Cmdy("web.cache", "watch", up[0], p)
	} else {
		// 下拉文件
		m.Cmdy("web.spide", "dev", "save", p, "GET",
			kit.MergeURL2(m.Option(MSG_USERWEB), path.Join("/share/cache", up[0])))
	}
}
func (m *Message) Action(arg ...string) {
	m.Option(MSG_ACTION, kit.Format(arg))
}
func (m *Message) Status(arg ...interface{}) {
	args := kit.Simple(arg)
	list := []map[string]interface{}{}
	for i := 0; i < len(args)-1; i += 2 {
		list = append(list, map[string]interface{}{
			kit.MDB_NAME: args[i], kit.MDB_VALUE: args[i+1],
		})
	}
	m.Option(MSG_STATUS, kit.Format(list))
}
func (m *Message) StatusTime(arg ...interface{}) {
	m.Status(kit.MDB_TIME, m.Time(), arg, "cost", m.FormatCost())
}
func (m *Message) StatusTimeCount(arg ...interface{}) {
	m.Status(kit.MDB_TIME, m.Time(), kit.MDB_COUNT, m.FormatSize(), arg, "cost", m.FormatCost())
}
func (m *Message) StatusTimeCountTotal(arg ...interface{}) {
	m.Status(kit.MDB_TIME, m.Time(), kit.MDB_COUNT, m.FormatSize(), "total", arg, "cost", m.FormatCost())
}

func (m *Message) Toast(content string, arg ...interface{}) {
	if len(arg) > 1 {
		switch val := arg[1].(type) {
		case string:
			if value, err := time.ParseDuration(val); err == nil {
				arg[1] = int(value / time.Millisecond)
			}
		}
	}

	if m.Option(MSG_USERPOD) == "" {
		m.Cmd("web.space", m.Option(MSG_DAEMON), "toast", "", content, arg)
	} else {
		m.Option(MSG_TOAST, kit.Simple(content, arg))
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
func (m *Message) ProcessCommand(cmd string, val []string, arg ...string) {
	if len(arg) > 0 && arg[0] == "run" {
		m.Cmdy(cmd, arg[1:])
		return
	}

	m.Cmdy("command", cmd)
	m.ProcessField(cmd, "run")
	m.Push("arg", kit.Format(val))
}
func (m *Message) ProcessCommandOpt(arg ...string) {
	m.Push("opt", kit.Format(m.OptionSimple(arg...)))
}
func (m *Message) ProcessField(arg ...interface{}) {
	m.Process(PROCESS_FIELD)
	m.Option("_prefix", arg...)
}
func (m *Message) ProcessDisplay(arg ...interface{}) {
	m.Process(PROCESS_DISPLAY)
	m.Option("_display", arg...)
}
func (m *Message) ProcessInner()          { m.Process(PROCESS_INNER) }
func (m *Message) ProcessAgain()          { m.Process(PROCESS_AGAIN) }
func (m *Message) ProcessHold()           { m.Process(PROCESS_HOLD) }
func (m *Message) ProcessBack()           { m.Process(PROCESS_BACK) }
func (m *Message) ProcessOpen(url string) { m.Process(PROCESS_OPEN, url) }

func (m *Message) ShowPlugin(pod, ctx, cmd string, arg ...string) {
	m.Cmdy("web.space", pod, "context", ctx, "command", cmd)
	m.Option(MSG_PROCESS, PROCESS_FIELD)
	m.Option(FIELD_PREFIX, arg)
}
func (m *Message) PushPodCmd(cmd string, arg ...string) {
	m.Table(func(index int, value map[string]string, head []string) {
		m.Push("pod", m.Option(MSG_USERPOD))
	})

	m.Cmd("web.space").Table(func(index int, value map[string]string, head []string) {
		switch value[kit.MDB_TYPE] {
		case "worker", "server":
			if value[kit.MDB_NAME] == Info.HostName {
				break
			}
			m.Cmd("web.space", value[kit.MDB_NAME], m.Prefix(cmd), arg).Table(func(index int, val map[string]string, head []string) {
				val["pod"] = kit.Keys(value[kit.MDB_NAME], val["pod"])
				m.Push("", val, head)
			})
		}
	})
}
func (m *Message) PushSearch(args ...interface{}) {
	data := kit.Dict(args...)
	for _, k := range kit.Split(m.Option(MSG_FIELDS)) {
		switch k {
		case "pod":
			// m.Push(k, kit.Select(m.Option(MSG_USERPOD), data[kit.SSH_POD]))
		case "ctx":
			m.Push(k, m.Prefix())
		case "cmd":
			m.Push(k, kit.Format(data["cmd"]))
		case kit.MDB_TIME:
			m.Push(k, kit.Select(m.Time(), data[k]))
		default:
			m.Push(k, kit.Format(kit.Select("", data[k])))
		}
	}
}
func (m *Message) PushSearchWeb(cmd string, name string) {
	msg := m.Spawn()
	msg.Option(MSG_FIELDS, "type,name,text")
	msg.Cmd("mdb.select", m.Prefix(cmd), "", kit.MDB_HASH).Table(func(index int, value map[string]string, head []string) {
		text := kit.MergeURL(value[kit.MDB_TEXT], value[kit.MDB_NAME], name)
		if value[kit.MDB_NAME] == "" {
			text = kit.MergeURL(value[kit.MDB_TEXT] + url.QueryEscape(name))
		}
		m.PushSearch("cmd", cmd, kit.MDB_TYPE, kit.Select("", value[kit.MDB_TYPE]), kit.MDB_NAME, name, kit.MDB_TEXT, text)
	})
}
