package ice

import (
	"net/http"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func AddRender(key string, render func(*Message, string, ...Any) string) {
	Info.render[key] = render
}
func Render(m *Message, cmd string, args ...Any) string {
	if render, ok := Info.render[cmd]; ok {
		return render(m, cmd, args...)
	}

	switch arg := kit.Simple(args...); cmd {
	case RENDER_ANCHOR: // [name] link
		p := kit.Select(arg[0], arg, 1)
		return kit.Format(`<a href="%s" target="_blank">%s</a>`, p, arg[0])

	case RENDER_BUTTON: // name...
		list := []string{}
		for _, k := range kit.Split(kit.Join(arg)) {
			list = append(list, kit.Format(`<input type="button" name="%s" value="%s">`, k,
				kit.Select(k, kit.Value(m._cmd.Meta, kit.Keys("_trans", k)), m.Option(MSG_LANGUAGE) != "en")))
		}
		return strings.Join(list, "")

	case RENDER_IMAGES: // src [height]
		m.Debug("what %v", m.Option(MSG_USERUA))
		if strings.Contains(m.Option(MSG_USERUA), "Mobile") {
			return kit.Format(`<img src="%s" width=%d>`, arg[0], kit.Int(kit.Select(kit.Select("120", m.Option("width")), arg, 1))-24)
		}
		return kit.Format(`<img src="%s" height=%d>`, arg[0], kit.Int(kit.Select(kit.Select("240", m.Option("height")), arg, 1))/2-24)

	case RENDER_VIDEOS: // src [height]
		return kit.Format(`<video src="%s" height=%s controls>`, arg[0], kit.Select("120", arg, 1))

	case RENDER_IFRAME: // src
		return kit.Format(`<iframe src="%s"></iframe>`, arg[0])
	default:
		return arg[0]
	}
	return ""
}

func (m *Message) Render(cmd string, args ...Any) *Message {
	switch cmd {
	case RENDER_TEMPLATE: // text [data]
		if len(args) == 1 {
			args = append(args, m)
		}
		if res, err := kit.Render(args[0].(string), args[1]); m.Assert(err) {
			m.Echo(string(res))
		}
		return m
	}
	m.Optionv(MSG_OUTPUT, cmd)
	m.Optionv(MSG_ARGS, args)

	return m
}
func (m *Message) RenderTemplate(args ...Any) *Message {
	return m.Render(RENDER_TEMPLATE, args...)
}
func (m *Message) RenderStatus(status int) *Message {
	return m.Render(RENDER_STATUS, status)
}
func (m *Message) RenderStatusBadRequest() *Message {
	return m.Render(RENDER_STATUS, http.StatusBadRequest)
}
func (m *Message) RenderStatusUnauthorized() *Message {
	return m.Render(RENDER_STATUS, http.StatusUnauthorized)
}
func (m *Message) RenderStatusForbidden() *Message {
	return m.Render(RENDER_STATUS, http.StatusForbidden)
}
func (m *Message) RenderStatusNotFound() *Message {
	return m.Render(RENDER_STATUS, http.StatusNotFound)
}
func (m *Message) RenderRedirect(args ...Any) *Message {
	return m.Render(RENDER_REDIRECT, args...)
}
func (m *Message) RenderDownload(args ...Any) *Message {

	m.Debug("what %v", kit.Format(args))
	return m.Render(RENDER_DOWNLOAD, args...)
}
func (m *Message) RenderResult(args ...Any) *Message { // [fmt arg...]
	return m.Render(RENDER_RESULT, args...)
}
func (m *Message) RenderJson(args ...Any) *Message { // [key value]...
	return m.Render(RENDER_JSON, kit.Format(kit.Dict(args...)))
}

func (m *Message) IsMobileUA() bool {
	return strings.Contains(m.Option(MSG_USERUA), "Mobile")
}
func (m *Message) IsCliUA() bool {
	if m.Option(MSG_USERUA) == "" || !strings.HasPrefix(m.Option(MSG_USERUA), "Mozilla") {
		return true
	}
	return false
}
func (m *Message) PushAction(list ...Any) *Message {
	if len(m.meta[MSG_APPEND]) == 0 {
		return m
	}
	return m.Set(MSG_APPEND, ACTION).Tables(func(value Maps) { m.PushButton(list...) })
}
func (m *Message) PushSearch(args ...Any) {
	data := kit.Dict(args...)
	for i := 0; i < len(args); i += 2 {
		switch k := args[i].(type) {
		case string:
			if i+1 < len(args) {
				data[k] = args[i+1]
			}
		}
	}
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
func (m *Message) PushAnchor(arg ...Any) { // [name] link
	if !m.IsCliUA() {
		m.Push(LINK, Render(m, RENDER_ANCHOR, arg...))
	}
}
func (m *Message) PushButton(arg ...Any) { // name...
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
func (m *Message) PushImages(key, src string, arg ...string) { // key src [height]
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_IMAGES, src, arg))
	}
}
func (m *Message) PushVideos(key, src string, arg ...string) { // key src [height]
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_VIDEOS, src, arg))
	}
}
func (m *Message) PushIFrame(key, src string, arg ...string) { // key src
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_IFRAME, src, arg))
	}
}
func (m *Message) PushQRCode(key string, src string, arg ...string) { // key src [height]
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_QRCODE, src, arg))
	}
}
func (m *Message) PushScript(arg ...string) { // [type] text...
	if !m.IsCliUA() {
		m.Push(SCRIPT, Render(m, RENDER_SCRIPT, arg))
	}
}
func (m *Message) PushDownload(key string, arg ...Any) { // [name] file
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_DOWNLOAD, arg...))
	}
}

func (m *Message) EchoAnchor(arg ...Any) *Message { // [name] link
	return m.Echo(Render(m, RENDER_ANCHOR, arg...))
}
func (m *Message) EchoButton(arg ...Any) *Message { // name...
	return m.Echo(Render(m, RENDER_BUTTON, arg...))
}
func (m *Message) EchoImages(src string, arg ...string) *Message { // src [height]
	return m.Echo(Render(m, RENDER_IMAGES, src, arg))
}
func (m *Message) EchoVideos(src string, arg ...string) *Message { // src [height]
	return m.Echo(Render(m, RENDER_VIDEOS, src, arg))
}
func (m *Message) EchoIFrame(src string, arg ...string) *Message { // src
	return m.Echo(Render(m, RENDER_IFRAME, src, arg))
}
func (m *Message) EchoQRCode(src string, arg ...string) *Message { // src [height]
	return m.Echo(Render(m, RENDER_QRCODE, src, arg))
}
func (m *Message) EchoScript(arg ...string) *Message { // [type] text...
	return m.Echo(Render(m, RENDER_SCRIPT, arg))
}
func (m *Message) EchoDownload(arg ...Any) *Message { // [name] file
	return m.Echo(Render(m, RENDER_DOWNLOAD, arg...))
}
