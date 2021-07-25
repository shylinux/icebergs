package ice

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	kit "github.com/shylinux/toolkits"
)

func (m *Message) Prefix(arg ...string) string {
	return kit.Keys(m.Cap(CTX_FOLLOW), arg)
}
func (m *Message) PrefixKey(arg ...string) string {
	return kit.Keys(m.Cap(CTX_FOLLOW), m._key, arg)
}
func (m *Message) Save(arg ...string) *Message {
	if len(arg) == 0 {
		for k := range m.target.Configs {
			arg = append(arg, k)
		}
	}
	list := []string{}
	for _, k := range arg {
		list = append(list, m.Prefix(k))
	}
	m.Cmd("ctx.config", "save", m.Prefix("json"), list)
	return m
}
func (m *Message) Load(arg ...string) *Message {
	list := []string{}
	for _, k := range arg {
		list = append(list, m.Prefix(k))
	}
	m.Cmd("ctx.config", "load", m.Prefix("json"), list)
	return m
}

func (m *Message) Watch(key string, arg ...string) *Message {
	if len(arg) == 0 {
		arg = append(arg, m.Prefix("auto"))
	}
	m.Cmd("gdb.event", "action", "listen", "event", key, kit.SSH_CMD, strings.Join(arg, " "))
	return m
}
func (m *Message) Event(key string, arg ...string) *Message {
	m.Cmd("gdb.event", "action", "action", "event", key, arg)
	return m
}
func (m *Message) Right(arg ...interface{}) bool {
	return m.Option(MSG_USERROLE) == "root" || !m.Warn(m.Cmdx("aaa.role", "right",
		m.Option(MSG_USERROLE), strings.ReplaceAll(kit.Keys(arg...), "/", ".")) != "ok",
		ErrNotRight, m.Option(MSG_USERROLE), " of ", strings.Join(kit.Simple(arg), "."), " at ", kit.FileLine(2, 3))
}
func (m *Message) Space(arg interface{}) []string {
	if arg == nil || arg == "" || kit.Format(arg) == m.Conf("cli.runtime", "node.name") {
		return nil
	}
	return []string{"web.space", kit.Format(arg)}
}

func (m *Message) ShowPlugin(pod, ctx, cmd string, arg ...string) {
	m.Cmdy("web.space", pod, "context", ctx, "command", cmd)
	m.Option(MSG_PROCESS, PROCESS_FIELD)
	m.Option(FIELD_PREFIX, arg)
}
func (m *Message) PushPodCmd(cmd string, arg ...string) {
	m.Table(func(index int, value map[string]string, head []string) {
		m.Push(kit.SSH_POD, m.Option(MSG_USERPOD))
	})

	m.Cmd("web.space").Table(func(index int, value map[string]string, head []string) {
		switch value[kit.MDB_TYPE] {
		case "worker", "server":
			if value[kit.MDB_NAME] == Info.HostName {
				break
			}
			m.Cmd("web.space", value[kit.MDB_NAME], m.Prefix(cmd), arg).Table(func(index int, val map[string]string, head []string) {
				val[kit.SSH_POD] = kit.Keys(value[kit.MDB_NAME], val[kit.SSH_POD])
				m.Push("", val, head)
			})
		}
	})
}
func (m *Message) PushSearch(args ...interface{}) {
	data := kit.Dict(args...)
	for _, k := range kit.Split(m.Option(MSG_FIELDS)) {
		switch k {
		case kit.SSH_POD:
			// m.Push(k, kit.Select(m.Option(MSG_USERPOD), data[kit.SSH_POD]))
		case kit.SSH_CTX:
			m.Push(k, m.Prefix())
		case kit.SSH_CMD:
			m.Push(k, data[kit.SSH_CMD])
		case kit.MDB_TIME:
			m.Push(k, kit.Select(m.Time(), data[k]))
		default:
			m.Push(k, kit.Select("", data[k]))
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
		m.PushSearch(kit.SSH_CMD, cmd, kit.MDB_TYPE, kit.Select("", value[kit.MDB_TYPE]), kit.MDB_NAME, name, kit.MDB_TEXT, text)
	})
}

func Render(m *Message, cmd string, args ...interface{}) string {
	if m.IsCliUA() {
		switch arg := kit.Simple(args...); cmd {
		case RENDER_QRCODE: // text [size]
			return m.Cmdx("cli.qrcode", arg[0])
		}
		return ""
	}

	switch arg := kit.Simple(args...); cmd {
	case RENDER_DOWNLOAD: // [name] file
		if arg[0] == "" {
			return ""
		}
		list := []string{}
		if m.Option(MSG_USERPOD) != "" {
			list = append(list, kit.SSH_POD, m.Option(MSG_USERPOD))
		}
		if len(arg) == 1 {
			arg[0] = kit.MergeURL2(m.Option(MSG_USERWEB), path.Join(kit.Select("", "/share/local",
				!strings.HasPrefix(arg[0], "/")), arg[0]), list)
		} else {
			arg[1] = kit.MergeURL2(m.Option(MSG_USERWEB), path.Join(kit.Select("", "/share/local",
				!strings.HasPrefix(arg[1], "/")), arg[1]), list, "filename", arg[0])
		}
		arg[0] = m.ReplaceLocalhost(arg[0])
		return fmt.Sprintf(`<a href="%s" download="%s">%s</a>`, m.ReplaceLocalhost(kit.Select(arg[0], arg, 1)), path.Base(arg[0]), arg[0])

	case RENDER_ANCHOR: // [name] link
		return fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, kit.Select(arg[0], arg, 1), arg[0])

	case RENDER_BUTTON: // name...
		if m._cmd == nil || m._cmd.Meta == nil {
			return ""
		}
		list := []string{}
		for _, k := range kit.Split(strings.Join(arg, ",")) {
			list = append(list, fmt.Sprintf(`<input type="button" name="%s" value="%s">`,
				k, kit.Select(k, kit.Value(m._cmd.Meta, kit.Keys("trans", k)))))
		}
		return strings.Join(list, "")

	case RENDER_IMAGES: // src [size]
		return fmt.Sprintf(`<img src="%s" height=%s>`, arg[0], kit.Select("120", arg, 1))

	case RENDER_VIDEOS: // src [size]
		return fmt.Sprintf(`<video src="%s" height=%s controls>`, arg[0], kit.Select("120", arg, 1))

	case RENDER_QRCODE: // text [size]
		return m.Cmdx("cli.qrcode", arg[0])

	case RENDER_SCRIPT: // type text
		if len(arg) == 1 && arg[0] != kit.SSH_BREAK {
			arg = []string{kit.SSH_SHELL, arg[0]}
		}
		list := []string{kit.Format(`<div class="story" data-type="spark" data-name="%s">`, arg[0])}
		for _, l := range strings.Split(strings.Join(arg[1:], "\n"), "\n") {
			switch list = append(list, "<div>"); arg[0] {
			case kit.SSH_SHELL:
				list = append(list, "<label>", "$ ", "</label>")
			default:
				list = append(list, "<label>", "&gt; ", "</label>")
			}
			list = append(list, "<span>", l, "</span>")
			list = append(list, "</div>")
		}
		list = append(list, "</div>")
		return strings.Join(list, "")
	default:
		return arg[0]
	}
	return ""
}
func (m *Message) PushRender(key, view, name string, arg ...string) *Message {
	return m.Push(key, Render(m, view, name, arg))
}
func (m *Message) PushDownload(key string, arg ...interface{}) { // [name] file
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_DOWNLOAD, arg...))
	}
}
func (m *Message) PushAnchor(arg ...interface{}) { // [name] link
	if !m.IsCliUA() {
		m.Push(kit.MDB_LINK, Render(m, RENDER_ANCHOR, arg...))
	}
}
func (m *Message) PushButton(arg ...string) {
	if !m.IsCliUA() {
		m.Push(kit.MDB_ACTION, Render(m, RENDER_BUTTON, strings.Join(arg, ",")))
	}
}
func (m *Message) PushScript(arg ...string) *Message { // [type] text...
	return m.Push(kit.MDB_SCRIPT, Render(m, RENDER_SCRIPT, arg))
}
func (m *Message) PushImages(key, src string, arg ...string) { // key src [size]
	m.Push(key, Render(m, RENDER_IMAGES, src, arg))
}
func (m *Message) PushVideos(key, src string, arg ...string) { // key src [size]
	m.Push(key, Render(m, RENDER_VIDEOS, src, arg))
}
func (m *Message) PushQRCode(key string, text string, arg ...string) { // key text [size]
	m.Push(key, Render(m, RENDER_QRCODE, text, arg))
}
func (m *Message) PushAction(list ...interface{}) {
	m.Table(func(index int, value map[string]string, head []string) {
		m.PushButton(kit.Simple(list...)...)
	})
}

func (m *Message) EchoAnchor(arg ...interface{}) *Message { // [name] link
	return m.Echo(Render(m, RENDER_ANCHOR, arg...))
}
func (m *Message) EchoButton(arg ...string) *Message {
	return m.Echo(Render(m, RENDER_BUTTON, strings.Join(arg, ",")))
}
func (m *Message) EchoScript(arg ...string) *Message {
	return m.Echo(Render(m, RENDER_SCRIPT, arg))
}
func (m *Message) EchoImages(src string, arg ...string) *Message {
	return m.Echo(Render(m, RENDER_IMAGES, src, arg))
}
func (m *Message) EchoQRCode(text string, arg ...string) *Message { // text [size]
	return m.Echo(Render(m, RENDER_QRCODE, text, arg))
}
func (m *Message) EchoDownload(arg ...interface{}) { // [name] file
	m.Echo(Render(m, RENDER_DOWNLOAD, arg...))
}

func (m *Message) SortInt(key string)   { m.Sort(key, "int") }
func (m *Message) SortIntR(key string)  { m.Sort(key, "int_r") }
func (m *Message) SortStr(key string)   { m.Sort(key, "str") }
func (m *Message) SortStrR(key string)  { m.Sort(key, "str_r") }
func (m *Message) SortTime(key string)  { m.Sort(key, "time") }
func (m *Message) SortTimeR(key string) { m.Sort(key, "time_r") }

func (m *Message) FormatMeta() string { return m.Format("meta") }
func (m *Message) FormatSize() string { return m.Format("size") }
func (m *Message) FormatCost() string { return m.Format("cost") }

func (m *Message) Render(cmd string, args ...interface{}) *Message {
	m.Optionv(MSG_OUTPUT, cmd)
	m.Optionv(MSG_ARGS, args)

	switch cmd {
	case RENDER_TEMPLATE: // text [data [type]]
		if len(args) == 1 {
			args = append(args, m)
		}
		if res, err := kit.Render(args[0].(string), args[1]); m.Assert(err) {
			m.Echo(string(res))
		}
	}
	return m
}
func (m *Message) RenderResult(args ...interface{}) *Message {
	return m.Render(RENDER_RESULT, args...)
}
func (m *Message) RenderTemplate(args ...interface{}) *Message {

	return m.Render(RENDER_TEMPLATE, args...)
}
func (m *Message) RenderDownload(args ...interface{}) *Message {
	return m.Render(RENDER_DOWNLOAD, args...)
}
func (m *Message) RenderRedirect(args ...interface{}) *Message {
	return m.Render(RENDER_REDIRECT, args...)
}
func (m *Message) RenderIndex(serve, repos string) *Message {
	return m.RenderDownload(path.Join(m.Conf(serve, kit.Keym(repos, kit.SSH_PATH)), m.Conf(serve, kit.Keym(repos, kit.SSH_INDEX))))
}

type Sort struct {
	Fields string
	Method string
}
type Option struct {
	Name  string
	Value interface{}
}

func (m *Message) OptionSimple(key ...string) (res []string) {
	for _, k := range strings.Split(strings.Join(key, ","), ",") {
		res = append(res, k, m.Option(k))
	}
	return
}
func OptionFields(str ...string) Option       { return Option{MSG_FIELDS, strings.Join(str, ",")} }
func OptionHash(str string) Option            { return Option{kit.MDB_HASH, str} }
func (m *Message) OptionFields(str ...string) { m.Option(MSG_FIELDS, strings.Join(str, ",")) }
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
	list := []map[string]string{}
	for i := 0; i < len(args)-1; i += 2 {
		list = append(list, map[string]string{
			"name": args[i], "value": args[i+1],
		})
	}
	m.Option(MSG_STATUS, kit.Format(list))
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
func (m *Message) ProcessRewrite(arg ...interface{}) {
	m.Process(PROCESS_REWRITE)
	m.Option("_arg", arg...)
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
func (m *Message) ProcessInner() { m.Process(PROCESS_INNER) }
func (m *Message) ProcessHold()  { m.Process(PROCESS_HOLD) }
func (m *Message) ProcessBack()  { m.Process(PROCESS_BACK) }

func (m *Message) Confi(key string, sub string) int {
	return kit.Int(m.Conf(key, sub))
}
func (m *Message) Capi(key string, val ...interface{}) int {
	if len(val) > 0 {
		m.Cap(key, kit.Int(m.Cap(key))+kit.Int(val[0]))
	}
	return kit.Int(m.Cap(key))
}
func (m *Message) Cut(fields ...string) *Message {
	m.meta[MSG_APPEND] = strings.Split(strings.Join(fields, ","), ",")
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
	fields := kit.Split(field, space, space, space)
	for i, l := range kit.Split(str, enter, enter, enter) {
		if strings.HasPrefix(l, "Binary") {
			continue
		}
		if strings.TrimSpace(l) == "" {
			continue
		}
		if i == 0 && (field == "" || field == "index") {
			// 表头行
			fields = kit.Split(l, space, space)
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

		ls := kit.Split(l, space, space)
		for i, v := range ls {
			if i == len(fields)-1 {
				m.Push(kit.Select("some", fields, i), strings.Join(ls[i:], space))
				break
			}
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
func (m *Message) RenameAppend(from, to string) {
	for i, v := range m.meta[MSG_APPEND] {
		if v == from {
			m.meta[MSG_APPEND][i] = to
			m.meta[to] = m.meta[from]
			delete(m.meta, from)
		}
	}
}
func (m *Message) IsCliUA() bool {
	if m.Option(MSG_USERUA) == "" || !strings.HasPrefix(m.Option(MSG_USERUA), "Mozilla/5.0") {
		return true
	}
	return false
}
func (m *Message) ReplaceLocalhost(url string) string {
	if strings.Contains(url, "://localhost") {
		return strings.Replace(url, "localhost", m.Cmd("tcp.host").Append("ip"), 1)
	}
	return url
}

func Display(file string, arg ...string) map[string]string {
	if file != "" && !strings.HasPrefix(file, "/") {
		ls := strings.Split(kit.FileLine(2, 100), "usr")
		file = kit.Select(file+".js", file, strings.HasSuffix(file, ".js"))
		file = path.Join("/require/github.com/shylinux", path.Dir(ls[len(ls)-1]), file)
	}
	// return map[string]string{kit.MDB_DISPLAY: file, kit.MDB_STYLE: kit.Select("", arg, 0)}
	return map[string]string{"display": file, kit.MDB_STYLE: kit.Select("", arg, 0)}
}
func (m *Message) OptionSplit(fields ...string) (res []string) {
	for _, k := range strings.Split(strings.Join(fields, ","), ",") {
		res = append(res, m.Option(k))
	}
	return res
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
