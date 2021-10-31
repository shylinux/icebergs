package ice

import (
	"path"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func AddRender(key string, render func(*Message, string, ...interface{}) string) {
	Info.render[key] = render
}
func Render(m *Message, cmd string, args ...interface{}) string {
	if render, ok := Info.render[cmd]; ok {
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
func (m *Message) RenderCmd(index string, args ...interface{}) {
	list := index
	if index != "" {
		msg := m.Cmd("command", index)
		list = kit.Format(kit.List(kit.Dict(
			kit.MDB_INDEX, index, kit.MDB_ARGS, kit.Simple(args),
			msg.AppendSimple(kit.MDB_NAME, kit.MDB_HELP),
			"feature", kit.UnMarshal(msg.Append("meta")),
			"inputs", kit.UnMarshal(msg.Append("list")),
		)))
	}
	m.RenderResult(kit.Format(`<!DOCTYPE html>
<head>
    <meta name="viewport" content="width=device-width,initial-scale=0.8,user-scalable=no">
    <meta charset="utf-8">
    <link rel="stylesheet" type="text/css" href="/page/cmd.css">
</head>
<body>
	<script src="/page/cmd.js"></script>
	<script>cmd(%s)</script>
</body>
`, list))
}

func (m *Message) IsCliUA() bool {
	if m.Option(MSG_USERUA) == "" || !strings.HasPrefix(m.Option(MSG_USERUA), "Mozilla/5.0") {
		return true
	}
	return false
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
func (m *Message) PushDownload(key string, arg ...interface{}) { // [name] file
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_DOWNLOAD, arg...))
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

	m.Cmd("space").Table(func(index int, value map[string]string, head []string) {
		switch value[kit.MDB_TYPE] {
		case "server", "worker":
			if value[kit.MDB_NAME] == Info.HostName {
				break
			}
			m.Cmd("space", value[kit.MDB_NAME], m.Prefix(cmd), arg).Table(func(index int, val map[string]string, head []string) {
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
			m.Push(k, kit.Select("", data[k]))
		case CTX:
			m.Push(k, kit.Select(m.Prefix(), data[k]))
		case CMD:
			m.Push(k, kit.Select(m.CommandKey(), data[k]))
		case kit.MDB_TIME:
			m.Push(k, kit.Select(m.Time(), data[k]))
		default:
			m.Push(k, kit.Select("", data[k]))
		}
	}
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
func (m *Message) EchoDownload(arg ...interface{}) *Message { // [name] file
	return m.Echo(Render(m, RENDER_DOWNLOAD, arg...))
}
