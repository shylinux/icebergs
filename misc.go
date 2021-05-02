package ice

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	kit "github.com/shylinux/toolkits"
)

func (m *Message) Prefix(arg ...string) string {
	return kit.Keys(m.Cap(CTX_FOLLOW), arg)
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
	for _, k := range kit.Split(m.Option("fields")) {
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
	msg.Option("fields", "type,name,text")
	msg.Cmd("mdb.select", m.Prefix(cmd), "", kit.MDB_HASH).Table(func(index int, value map[string]string, head []string) {
		text := kit.MergeURL(value[kit.MDB_TEXT], value[kit.MDB_NAME], name)
		if value[kit.MDB_NAME] == "" {
			text = kit.MergeURL(value[kit.MDB_TEXT] + url.QueryEscape(name))
		}
		m.PushSearch(kit.SSH_CMD, cmd, kit.MDB_TYPE, kit.Select("", value[kit.MDB_TYPE]), kit.MDB_NAME, name, kit.MDB_TEXT, text)
	})
}

func (m *Message) IsTermUA() bool {
	return m.Option(MSG_USERUA) == "" || strings.Contains(m.Option(MSG_USERUA), "curl")
}

func Render(m *Message, cmd string, args ...interface{}) string {
	if m.IsTermUA() {
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
		if len(arg) == 1 {
			arg[0] = kit.MergeURL2(m.Option(MSG_USERWEB), path.Join(kit.Select("", "/share/local",
				!strings.HasPrefix(arg[0], "/")), arg[0]), kit.SSH_POD, m.Option(MSG_USERPOD))
		} else {
			arg[1] = kit.MergeURL2(m.Option(MSG_USERWEB), path.Join(kit.Select("", "/share/local",
				!strings.HasPrefix(arg[1], "/")), arg[1]), kit.SSH_POD, m.Option(MSG_USERPOD))
		}
		return fmt.Sprintf(`<a href="%s" download="%s">%s</a>`, kit.Select(arg[0], arg, 1), path.Base(arg[0]), arg[0])

	case RENDER_ANCHOR: // [name] link
		return fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, kit.Select(arg[0], arg, 1), arg[0])

	case RENDER_BUTTON: // name...
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
	m.Push(key, Render(m, RENDER_DOWNLOAD, arg...))
}
func (m *Message) PushAnchor(arg ...interface{}) { // [name] link
	if m.IsTermUA() {
		return
	}
	m.Push(kit.MDB_LINK, Render(m, RENDER_ANCHOR, arg...))
}
func (m *Message) PushButton(arg ...string) {
	if m.IsTermUA() {
		return
	}
	m.Push(kit.MDB_ACTION, Render(m, RENDER_BUTTON, strings.Join(arg, ",")))
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

func (m *Message) SortInt(key string)   { m.Sort(key, "int") }
func (m *Message) SortIntR(key string)  { m.Sort(key, "int_r") }
func (m *Message) SortStr(key string)   { m.Sort(key, "str") }
func (m *Message) SortStrR(key string)  { m.Sort(key, "str_r") }
func (m *Message) SortTime(key string)  { m.Sort(key, "time") }
func (m *Message) SortTimeR(key string) { m.Sort(key, "time_r") }

func (m *Message) FormatMeta() string { return m.Format("meta") }
func (m *Message) RenameAppend(from, to string) {
	for i, v := range m.meta[MSG_APPEND] {
		if v == from {
			m.meta[MSG_APPEND][i] = to
			m.meta[to] = m.meta[from]
			delete(m.meta, from)
		}
	}
}

type Option struct {
	Name  string
	Value interface{}
}

func OptionFields(str string) Option { return Option{"fields", str} }
func OptionHash(str string) Option   { return Option{kit.MDB_HASH, str} }

func (m *Message) Toast(content string, arg ...interface{}) {
	m.Cmd("web.space", m.Option("_daemon"), "toast", "", content, arg)
}
func (m *Message) GoToast(title string, cb func(func(string, int, int))) {
	m.Go(func() {
		cb(func(name string, count, total int) {
			m.Toast(
				kit.Format("%s %d/%d", name, count, total),
				kit.Format("%s %d%%", title, count*100/total),
				kit.Select("1000", "10000", count < total),
				count*100/total,
			)
		})
	})
}

func (m *Message) Fields(condition bool, fields string) string {
	return m.Option("fields", kit.Select(kit.Select("detail", fields, condition), m.Option("fields")))
}
func (m *Message) Action(arg ...string) {
	m.Option(MSG_ACTION, kit.Format(arg))
}
func (m *Message) Process(action string, arg ...interface{}) {
	m.Option(MSG_PROCESS, action)
	m.Option("_arg", arg...)
}
func (m *Message) ProcessHold() { m.Process(PROCESS_HOLD) }
func (m *Message) ProcessBack() { m.Process(PROCESS_BACK) }
