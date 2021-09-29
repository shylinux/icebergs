package ice

import (
	"path"
	"reflect"
	"strings"

	kit "shylinux.com/x/toolkits"
)

var renderList = map[string]func(*Message, string, ...interface{}) string{}

func AddRender(key string, render func(*Message, string, ...interface{}) string) {
	renderList[key] = render
}
func Render(m *Message, cmd string, args ...interface{}) string {
	if render, ok := renderList[cmd]; ok {
		m.Debug("render: %v %v", cmd, kit.FileLine(render, 3))
		return render(m, cmd, args...)
	}

	switch arg := kit.Simple(args...); cmd {
	case RENDER_ANCHOR: // [name] link
		return kit.Format(`<a href="%s" target="_blank">%s</a>`, kit.Select(arg[0], arg, 1), arg[0])

	case RENDER_BUTTON: // name...
		if m._cmd == nil || m._cmd.Meta == nil {
			break
		}
		list := []string{}
		for _, k := range kit.Split(kit.Join(arg)) {
			list = append(list, kit.Format(`<input type="button" name="%s" value="%s">`,
				k, kit.Select(k, kit.Value(m._cmd.Meta, kit.Keys("_trans", k)), m.Option("language") != "en")))
		}
		return kit.Join(list, " ")

	case RENDER_IMAGES: // src [size]
		return kit.Format(`<img src="%s" height=%s>`, arg[0], kit.Select("120", arg, 1))

	case RENDER_VIDEOS: // src [size]
		return kit.Format(`<video src="%s" height=%s controls>`, arg[0], kit.Select("120", arg, 1))

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
		m.Push(kit.MDB_ACTION, Render(m, RENDER_BUTTON, strings.ToLower(kit.Join(arg))))
	}
}
func (m *Message) PushScript(arg ...string) *Message { // [type] text...
	return m.Push(kit.MDB_SCRIPT, Render(m, RENDER_SCRIPT, arg))
}
func (m *Message) PushQRCode(key string, text string, arg ...string) { // key text [size]
	m.Push(key, Render(m, RENDER_QRCODE, text, arg))
}
func (m *Message) PushImages(key, src string, arg ...string) { // key src [size]
	m.Push(key, Render(m, RENDER_IMAGES, src, arg))
}
func (m *Message) PushVideos(key, src string, arg ...string) { // key src [size]
	m.Push(key, Render(m, RENDER_VIDEOS, src, arg))
}
func (m *Message) PushAction(list ...interface{}) {
	for i, item := range list {
		if t := reflect.TypeOf(item); t.Kind() == reflect.Func {
			list[i] = kit.FuncName(item)
		}
	}

	m.Table(func(index int, value map[string]string, head []string) {
		m.PushButton(kit.Simple(list...)...)
	})
}

func (m *Message) EchoDownload(arg ...interface{}) { // [name] file
	m.Echo(Render(m, RENDER_DOWNLOAD, arg...))
}
func (m *Message) EchoAnchor(arg ...interface{}) *Message { // [name] link
	return m.Echo(Render(m, RENDER_ANCHOR, arg...))
}
func (m *Message) EchoButton(arg ...string) *Message {
	return m.Echo(Render(m, RENDER_BUTTON, kit.Join(arg)))
}
func (m *Message) EchoScript(arg ...string) *Message {
	return m.Echo(Render(m, RENDER_SCRIPT, arg))
}
func (m *Message) EchoQRCode(text string, arg ...string) *Message { // text [size]
	return m.Echo(Render(m, RENDER_QRCODE, text, arg))
}
func (m *Message) EchoImages(src string, arg ...string) *Message {
	return m.Echo(Render(m, RENDER_IMAGES, src, arg))
}

func (m *Message) IsCliUA() bool {
	if m.Option(MSG_USERUA) == "" || !strings.HasPrefix(m.Option(MSG_USERUA), "Mozilla/5.0") {
		return true
	}
	return false
}

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
func (m *Message) RenderIndex(serve, repos string, file ...string) *Message {
	return m.RenderDownload(path.Join(m.Conf(serve, kit.Keym(repos, kit.SSH_PATH)), kit.Select(m.Conf(serve, kit.Keym(repos, kit.SSH_INDEX)), path.Join(file...))))
}
