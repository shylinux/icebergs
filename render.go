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
		return render(m, cmd, args...)
	}

	switch arg := kit.Simple(args...); cmd {
	case RENDER_ANCHOR: // [name] link
		p := kit.Select(arg[0], arg, 1)
		if !strings.HasPrefix(p, HTTP) {
			p = m.MergeURL2(p)
		}
		return kit.Format(`<a href="%s" target="_blank">%s</a>`, p, arg[0])

	case RENDER_BUTTON: // name...
		if m._cmd == nil || m._cmd.Meta == nil {
			break
		}
		list := []string{}
		for _, k := range kit.Split(kit.Join(arg)) {
			list = append(list, kit.Format(`<input type="button" name="%s" value="%s">`, k,
				kit.Select(k, kit.Value(m._cmd.Meta, kit.Keys("_trans", k)), m.Option(MSG_LANGUAGE) != "en")))
		}
		return kit.Join(list, SP)

	case RENDER_IMAGES: // src [size]
		return kit.Format(`<img src="%s" height=%s>`, arg[0], kit.Select("120", arg, 1))

	case RENDER_VIDEOS: // src [size]
		return kit.Format(`<video src="%s" height=%s controls>`, arg[0], kit.Select("120", arg, 1))

	case RENDER_IFRAME: // src [size]
		return kit.Format(`<iframe src="%s"></iframe>`, arg[0])
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
func (m *Message) RenderWebsite(pod string, dir string, arg ...string) *Message {
	m.Cmdy("space", pod, "website", "action", "show", dir, arg)
	return m.RenderResult()
}
func (m *Message) RenderIndex(serve, repos string, file ...string) *Message {
	return m.RenderDownload(path.Join(m.Conf(serve, kit.Keym(repos, "path")), kit.Select(m.Conf(serve, kit.Keym(repos, INDEX)), path.Join(file...))))
}
func (m *Message) RenderCmd(index string, args ...interface{}) {
	list := index
	if index != "" {
		msg := m.Cmd(COMMAND, index)
		list = kit.Format(kit.List(kit.Dict(
			INDEX, index, ARGS, kit.Simple(args),
			msg.AppendSimple(NAME, HELP),
			FEATURE, kit.UnMarshal(msg.Append(META)),
			INPUTS, kit.UnMarshal(msg.Append(LIST)),
		)))
	}
	m.RenderResult(kit.Format(`<!DOCTYPE html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width,initial-scale=0.8,user-scalable=no">
    <link rel="stylesheet" type="text/css" href="/page/can.css">
	<script src="/page/can.js"></script>
	<script>can(%s)</script>
</head>
<body>
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
		m.Push(LINK, Render(m, RENDER_ANCHOR, arg...))
	}
}
func (m *Message) PushButton(arg ...interface{}) { // name...
	if !m.IsCliUA() {
		if m.FieldsIsDetail() {
			for i, k := range m.meta[KEY] {
				if k == ACTION {
					m.meta[VALUE][i] = Render(m, RENDER_BUTTON, arg...)
					return
				}
			}
		} else if len(m.meta[ACTION]) >= m.Length() {
			m.meta[ACTION] = []string{}
		}
		m.Push(ACTION, Render(m, RENDER_BUTTON, arg...))
	}
}
func (m *Message) PushScript(arg ...string) { // [type] text...
	if !m.IsCliUA() {
		m.Push(SCRIPT, Render(m, RENDER_SCRIPT, arg))
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
func (m *Message) PushIFrame(key, src string, arg ...string) { // key src [size]
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_IFRAME, src, arg))
	}
}
func (m *Message) PushDownload(key string, arg ...interface{}) { // [name] file
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_DOWNLOAD, arg...))
	}
}

func (m *Message) PushAction(list ...interface{}) *Message {
	if len(m.meta[MSG_APPEND]) == 0 {
		return m
	}
	return m.Set(MSG_APPEND, ACTION).Table(func(index int, value map[string]string, head []string) {
		m.PushButton(list...)
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
		case TIME:
			m.Push(k, kit.Select(m.Time(), data[k]))
		default:
			m.Push(k, kit.Select("", data[k]))
		}
	}
}
func (m *Message) PushPodCmd(cmd string, arg ...string) {
	if m.Length() > 0 && len(m.Appendv(POD)) == 0 {
		m.Table(func(index int, value map[string]string, head []string) {
			m.Push(POD, m.Option(MSG_USERPOD))
		})
	}

	m.Cmd(SPACE).Table(func(index int, value map[string]string, head []string) {
		switch value[TYPE] {
		case "server", "worker":
			if value[NAME] == Info.HostName {
				break
			}
			m.Cmd(SPACE, value[NAME], m.Prefix(cmd), arg).Table(func(index int, val map[string]string, head []string) {
				val[POD] = kit.Keys(value[NAME], val[POD])
				m.Push("", val, head)
			})
		}
	})
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
func (m *Message) EchoIFrame(src string, arg ...string) *Message { // src [size]
	return m.Echo(Render(m, RENDER_IFRAME, src, arg))
}
func (m *Message) EchoDownload(arg ...interface{}) *Message { // [name] file
	return m.Echo(Render(m, RENDER_DOWNLOAD, arg...))
}

func (m *Message) DisplayBase(file string, arg ...interface{}) *Message {
	if !strings.Contains(file, PT) {
		file += ".js"
	}
	m.Option(MSG_DISPLAY, kit.MergeURL(DisplayBase(file)[DISPLAY], arg...))
	return m
}
func (m *Message) DisplayStory(file string, arg ...interface{}) *Message {
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PLUGIN_STORY, file)
	}
	return m.DisplayBase(file, arg...)
}
func (m *Message) DisplayLocal(file string, arg ...interface{}) *Message {
	if file == "" {
		file = path.Join(kit.PathName(2), kit.Keys(kit.FileName(2), JS))
	}
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PLUGIN_LOCAL, file)
	}
	return m.DisplayBase(file, arg...)
}
func (m *Message) Display(file string, arg ...interface{}) *Message {
	m.Option(MSG_DISPLAY, kit.MergeURL(DisplayRequire(2, file)[DISPLAY], arg...))
	return m
}
func (m *Message) DisplayStoryJSON(arg ...interface{}) *Message {
	return m.DisplayStory("json", arg...)
}

func DisplayBase(file string, arg ...string) map[string]string {
	return map[string]string{DISPLAY: file, STYLE: kit.Join(arg, SP)}
}
func DisplayStory(file string, arg ...string) map[string]string {
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PLUGIN_STORY, file)
	}
	return DisplayBase(file, arg...)
}
func DisplayLocal(file string, arg ...string) map[string]string {
	if file == "" {
		file = path.Join(kit.PathName(2), kit.Keys(kit.FileName(2), JS))
	}
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PLUGIN_LOCAL, file)
	}
	return DisplayBase(file, arg...)
}
func DisplayRequire(n int, file string, arg ...string) map[string]string {
	if file == "" {
		file = kit.Keys(kit.FileName(n+1), JS)
	}
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		if kit.FileExists("src/" + file) {
			file = path.Join(PS, REQUIRE, "src/", file)
		} else {
			file = path.Join(PS, REQUIRE, kit.ModPath(n+1, file))
		}
	}
	return DisplayBase(file, arg...)
}
func Display(file string, arg ...string) map[string]string {
	return DisplayRequire(2, file, arg...)
}
