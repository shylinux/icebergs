package ice

import (
	"net/url"
	"path"
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
		for _, k := range kit.Split(strings.ToLower(kit.Join(arg))) {
			list = append(list, kit.Format(`<input type="button" name="%s" value="%s">`,
				k, kit.Select(k, kit.Value(m._cmd.Meta, kit.Keys("_trans", k)), m.Option(MSG_LANGUAGE) != "en")))
		}
		return kit.Join(list, SP)

	case RENDER_IMAGES: // src [size]
		return kit.Format(`<img src="%s" height=%s>`, arg[0], kit.Select("120", arg, 1))

	case RENDER_VIDEOS: // src [size]
		return kit.Format(`<video src="%s" height=%s controls>`, arg[0], kit.Select("120", arg, 1))

	default:
		return arg[0]
	}
	return ""
}

func (m *Message) IsCliUA() bool {
	if m.Option(MSG_USERUA) == "" || !strings.HasPrefix(m.Option(MSG_USERUA), "Mozilla/5.0") {
		return true
	}
	return false
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
func (m *Message) PushButton(arg ...interface{}) { // name...
	if !m.IsCliUA() {
		m.Push(kit.MDB_ACTION, Render(m, RENDER_BUTTON, arg...))
	}
}
func (m *Message) PushScript(arg ...string) { // [type] text...
	if !m.IsCliUA() {
		m.Push(kit.MDB_SCRIPT, Render(m, RENDER_SCRIPT, arg))
	}
}
func (m *Message) PushQRCode(key string, src string, arg ...string) { // key src [size]
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_QRCODE, src, arg))
	}
}
func (m *Message) PushImages(key, src string, arg ...string) { // key src [size]
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_IMAGES, src, arg))
	}
}
func (m *Message) PushVideos(key, src string, arg ...string) { // key src [size]
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_VIDEOS, src, arg))
	}
}

func (m *Message) PushAction(list ...interface{}) {
	m.Table(func(index int, value map[string]string, head []string) {
		m.PushButton(list...)
	})
}
func (m *Message) PushPodCmd(cmd string, arg ...string) {
	if strings.Contains(m.OptionFields(), POD) {
		m.Table(func(index int, value map[string]string, head []string) {
			m.Push(POD, m.Option(MSG_USERPOD))
		})
	}

	m.Cmd("web.space").Table(func(index int, value map[string]string, head []string) {
		switch value[kit.MDB_TYPE] {
		case "worker", "server":
			if value[kit.MDB_NAME] == Info.HostName {
				break
			}
			m.Cmd("web.space", value[kit.MDB_NAME], m.Prefix(cmd), arg).Table(func(index int, val map[string]string, head []string) {
				val[POD] = kit.Keys(value[kit.MDB_NAME], val[POD])
				m.Push("", val, head)
			})
		}
	})
}
func (m *Message) PushSearch(args ...interface{}) {
	data := kit.Dict(args...)
	for _, k := range kit.Split(m.OptionFields()) {
		switch k {
		case POD:
			m.Push(k, "")
			// m.Push(k, kit.Select(m.Option(MSG_USERPOD), data[kit.SSH_POD]))
		case CTX:
			m.Push(k, m.Prefix())
		case CMD:
			m.Push(k, kit.Format(data[CMD]))
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
		m.PushSearch(CMD, cmd, kit.MDB_TYPE, kit.Select("", value[kit.MDB_TYPE]), kit.MDB_NAME, name, kit.MDB_TEXT, text)
	})
}

func (m *Message) EchoDownload(arg ...interface{}) *Message { // [name] file
	return m.Echo(Render(m, RENDER_DOWNLOAD, arg...))
}
func (m *Message) EchoAnchor(arg ...interface{}) *Message { // [name] link
	return m.Echo(Render(m, RENDER_ANCHOR, arg...))
}
func (m *Message) EchoButton(arg ...interface{}) *Message { // name...
	return m.Echo(Render(m, RENDER_BUTTON, arg...))
}
func (m *Message) EchoScript(arg ...string) *Message { // [type] text...
	return m.Echo(Render(m, RENDER_SCRIPT, arg))
}
func (m *Message) EchoQRCode(src string, arg ...string) *Message { // src [size]
	return m.Echo(Render(m, RENDER_QRCODE, src, arg))
}
func (m *Message) EchoImages(src string, arg ...string) *Message { // src [size]
	return m.Echo(Render(m, RENDER_IMAGES, src, arg))
}
func (m *Message) EchoVideos(src string, arg ...string) *Message { // src [size]
	return m.Echo(Render(m, RENDER_VIDEOS, src, arg))
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
func (m *Message) RenderRedirect(args ...interface{}) *Message {
	return m.Render(RENDER_REDIRECT, args...)
}
func (m *Message) RenderDownload(args ...interface{}) *Message {
	return m.Render(RENDER_DOWNLOAD, args...)
}
func (m *Message) RenderIndex(serve, repos string, file ...string) *Message {
	return m.RenderDownload(path.Join(m.Conf(serve, kit.Keym(repos, kit.MDB_PATH)), kit.Select(m.Conf(serve, kit.Keym(repos, kit.MDB_INDEX)), path.Join(file...))))
}
