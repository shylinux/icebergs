package ice

import (
	kit "github.com/shylinux/toolkits"

	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/skip2/go-qrcode"
	"net/url"
	"path"
	"strings"
	"sync/atomic"
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
	m.Cmd("gdb.event", "action", "listen", "event", key, "cmd", strings.Join(arg, " "))
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

func (m *Message) PushPlugins(pod, ctx, cmd string, arg ...string) {
	m.Cmdy("space", pod, "context", ctx, "command", cmd)
	m.Option(MSG_PROCESS, PROCESS_FIELD)
	m.Option(FIELD_PREFIX, arg)
}
func (m *Message) PushPlugin(key string, arg ...string) {
	m.Cmdy("command", key)
	m.Option(MSG_PROCESS, PROCESS_FIELD)
	m.Option(FIELD_PREFIX, arg)
}
func (m *Message) PushRender(key, view, name string, arg ...string) *Message {
	return m.Push(key, _render(m, view, name, arg))
}
func (m *Message) PushPodCmd(cmd string, arg ...string) {
	m.Table(func(index int, value map[string]string, head []string) {
		m.Push("pod", m.Option(MSG_USERPOD))
	})

	m.Cmd("web.space").Table(func(index int, value map[string]string, head []string) {
		switch value["type"] {
		case "worker", "server":
			m.Cmd("web.space", value["name"], m.Prefix(cmd), arg).Table(func(index int, val map[string]string, head []string) {
				val["pod"] = kit.Keys(value["name"], val["pod"])
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
			m.Push(k, kit.Select(m.Option(MSG_USERPOD), data[kit.SSH_POD]))
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
	msg.Cmd("mdb.select", m.Prefix(cmd), "", "hash").Table(func(index int, value map[string]string, head []string) {
		text := kit.MergeURL(value["text"], value["name"], name)
		if value["name"] == "" {
			text = kit.MergeURL(value["text"] + url.QueryEscape(name))
		}
		m.PushSearch("cmd", cmd, "type", kit.Select("", value["type"]), "name", name, "text", text)
	})
}

func _render(m *Message, cmd string, args ...interface{}) string {
	if m.Option(MSG_USERUA) == "" || strings.Contains(m.Option(MSG_USERUA), "curl") {
		return ""
	}

	switch arg := kit.Simple(args...); cmd {
	case RENDER_DOWNLOAD: // [name] link
		if len(arg) == 1 {
			arg[0] = kit.MergeURL2(m.Option(MSG_USERWEB), path.Join("/share/local", arg[0]), "pod", m.Option(MSG_USERPOD))
		} else {
			arg[1] = kit.MergeURL2(m.Option(MSG_USERWEB), path.Join("/share/local", arg[1]), "pod", m.Option(MSG_USERPOD))
		}
		return fmt.Sprintf(`<a href="%s" download="%s">%s</a>`, kit.Select(arg[0], arg, 1), path.Base(arg[0]), arg[0])

	case RENDER_ANCHOR: // [name] link
		return fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, kit.Select(arg[0], arg, 1), arg[0])

	case RENDER_BUTTON: // name...
		list := []string{}
		for _, k := range kit.Split(strings.Join(arg, ",")) {
			list = append(list, fmt.Sprintf(`<input type="button" name="%s" value="%s">`,
				k, kit.Select(k, kit.Value(m.cmd.Meta, kit.Keys("trans", k)))))
		}
		return strings.Join(list, "")

	case RENDER_IMAGES: // src size
		return fmt.Sprintf(`<img src="%s" height=%s>`, arg[0], kit.Select("120", arg, 1))

	case RENDER_QRCODE: // text size height
		buf := bytes.NewBuffer(make([]byte, 0, MOD_BUFS))
		if qr, e := qrcode.New(arg[0], qrcode.Medium); m.Assert(e) {
			m.Assert(qr.Write(kit.Int(kit.Select("240", arg, 1)), buf))
		}
		src := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
		return fmt.Sprintf(`<img src="%s" title='%s' height=%s>`, src, arg[0], kit.Select("240", arg, 1))

	case RENDER_SCRIPT: // type text
		if arg[1] == "break" {
			return "<br>"
		}
		list := []string{}
		list = append(list, kit.Format(`<div class="story" data-type="spark" data-name="%s">`, arg[0]))
		for _, l := range strings.Split(arg[1], "\n") {
			list = append(list, "<div>")
			switch arg[0] {
			case "shell":
				list = append(list, "<label>$ </label>")
			default:
				list = append(list, "<label>&lt; </label>")
			}
			list = append(list, "<span>")
			list = append(list, l)
			list = append(list, "</span>")
			list = append(list, "</div>")
		}
		list = append(list, "</div>")
		return strings.Join(list, "")
	}
	return ""
}
func (m *Message) PushDownload(arg ...interface{}) { // [name] link
	m.Push("link", _render(m, RENDER_DOWNLOAD, arg...))
}
func (m *Message) PushAnchor(arg ...interface{}) { // [name] link
	m.Push("link", _render(m, RENDER_ANCHOR, arg...))
}
func (m *Message) PushButton(arg ...string) {
	m.Push("action", _render(m, RENDER_BUTTON, strings.Join(arg, ",")))
}
func (m *Message) PushScript(text string, arg ...string) *Message {
	mime := "shell"
	if len(arg) > 0 {
		mime, text = text, strings.Join(arg, "\n")
	}
	return m.Push("script", _render(m, RENDER_SCRIPT, mime, text))
}
func (m *Message) PushImages(key, src string, arg ...string) { // src [size]
	m.Push(key, _render(m, RENDER_IMAGES, src, arg))
}
func (m *Message) PushQRCode(key string, text string, arg ...string) { // text [size]
	m.Push(key, _render(m, RENDER_QRCODE, text, arg))
}
func (m *Message) PushAction(list ...interface{}) {
	m.Table(func(index int, value map[string]string, head []string) {
		m.PushButton(kit.Simple(list...)...)
	})
}

func (m *Message) EchoAnchor(arg ...interface{}) *Message { // [name] link
	return m.Echo(_render(m, RENDER_ANCHOR, arg...))
}
func (m *Message) EchoButton(arg ...string) *Message {
	return m.Echo(_render(m, RENDER_BUTTON, strings.Join(arg, ",")))
}
func (m *Message) EchoScript(text string, arg ...string) *Message {
	mime := "shell"
	if len(arg) > 0 {
		mime, text = text, strings.Join(arg, "\n")
	}
	return m.Echo(_render(m, RENDER_SCRIPT, mime, text))
}
func (m *Message) EchoQRCode(text string, arg ...string) *Message { // text [size]
	return m.Echo(_render(m, RENDER_QRCODE, text, arg))
}

func (m *Message) SortStr(key string)   { m.Sort(key, "str") }
func (m *Message) SortStrR(key string)  { m.Sort(key, "str_r") }
func (m *Message) SortInt(key string)   { m.Sort(key, "int") }
func (m *Message) SortIntR(key string)  { m.Sort(key, "int_r") }
func (m *Message) SortTime(key string)  { m.Sort(key, "time") }
func (m *Message) SortTimeR(key string) { m.Sort(key, "time_r") }

func (m *Message) PushRenderOld(key, view, name string, arg ...string) *Message {
	switch view {
	case "button":
		list := []string{}
		for _, k := range kit.Split(name) {
			list = append(list, fmt.Sprintf(`<input type="button" name="%s" value="%s">`,
				k, kit.Select(k, kit.Value(m.cmd.Meta, kit.Keys("trans", k)))))
		}
		m.Push(key, strings.Join(list, ""))
	case "video":
		m.Push(key, fmt.Sprintf(`<video src="%s" height=%s controls>`, name, kit.Select("120", arg, 0)))
	case "img":
		m.Push(key, fmt.Sprintf(`<img src="%s" height=%s>`, name, kit.Select("120", arg, 0)))
	case "a": // name [link]
		m.Push(key, fmt.Sprintf(`<a href="%s" target="_blank">%s</a>`, kit.Select(name, arg, 0), name))
	case "download": // name [link]
		m.Push(key, fmt.Sprintf(`<a href="%s" download="%s">%s</a>`, kit.Select(name, arg, 0), path.Base(name), name))
	default:
		m.Push(key, name)
	}
	return m
}

var count = int32(0)

func (m *Message) AddCmd(cmd *Command) string {
	name := fmt.Sprintf("_cb_%d", atomic.AddInt32(&count, 1))
	m.target.Commands[name] = cmd
	return kit.Keys(m.target.Cap(CTX_FOLLOW), name)
}
