package ice

import (
	"net/http"
	"strings"

	kit "shylinux.com/x/toolkits"
)

func AddRender(key string, render func(*Message, ...Any) string) {
	Info.render[key] = render
}
func RenderAction(key ...string) Actions {
	return Actions{CTX_INIT: {Hand: func(m *Message, arg ...string) {
		cmd := m.CommandKey()
		for _, key := range key {
			AddRender(key, func(m *Message, arg ...Any) string { return m.Cmd(cmd, key, arg).Result() })
		}
	}}}
}
func Render(m *Message, cmd string, args ...Any) string {
	if render, ok := Info.render[cmd]; ok {
		return render(m, args...)
	}
	switch arg := kit.Simple(args...); cmd {
	case RENDER_BUTTON:
		list := []string{}
		for _, k := range args {
			switch k := k.(type) {
			case []string:
				for _, k := range k {
					list = append(list, Render(m, RENDER_BUTTON, k))
				}
			case string:
				if strings.HasPrefix(k, "<input") {
					list = append(list, k)
					break
				}
				for _, k := range kit.Split(k) {
					list = append(list, kit.Format(`<input type="button" name="%s" value="%s">`, k,
						kit.Select(k, kit.Value(m._cmd.Meta, kit.Keys("_trans", k)), m.Option(MSG_LANGUAGE) != "en")))
				}
			case Map:
				for k, v := range k {
					list = append(list, kit.Format(`<input type="button" name="%s" value="%s">`, k, kit.Select(k, v, m.Option(MSG_LANGUAGE) != "en")))
				}
			}
		}
		return strings.Join(list, "")
	case RENDER_ANCHOR:
		return kit.Format(`<a href="%s" target="_blank">%s</a>`, kit.Select(arg[0], arg, 1), arg[0])
	case RENDER_IMAGES:
		return kit.Format(`<img src="%s" style="max-height:%spx; max-width:%spx">`, arg[0], m.Option(HEIGHT), m.Option(WIDTH))
	case RENDER_VIDEOS:
		return kit.Format(`<video src="%s" style="max-height:%spx; max-width:%spx" controls>`, arg[0], m.Option(HEIGHT), m.Option(WIDTH))
	case RENDER_IFRAME:
		return kit.Format(`<iframe src="%s"></iframe>`, arg[0])
	case RENDER_SCRIPT:
		return kit.Format(`<code>%s</code>`, arg[0])
	default:
		if len(arg) == 1 {
			return kit.Format(`<%s>%s</%s>`, cmd, arg[0], cmd)
		}
		return kit.Format(`<%s style="%s">%s</%s>`, cmd, kit.JoinKV(":", ";", arg[1:]...), arg[0], cmd)
	}
}

func (m *Message) Render(cmd string, arg ...Any) *Message {
	switch cmd {
	case RENDER_TEMPLATE:
		if len(arg) == 1 {
			arg = append(arg, m)
		}
		if res, err := kit.Render(arg[0].(string), arg[1]); m.Assert(err) {
			m.Echo(string(res))
		}
		return m
	}
	m.Options(MSG_OUTPUT, cmd, MSG_ARGS, arg)
	return m
}
func (m *Message) RenderTemplate(arg ...Any) *Message {
	return m.Render(RENDER_TEMPLATE, arg...)
}
func (m *Message) RenderStatus(status int, arg ...string) *Message {
	return m.Render(RENDER_STATUS, status, arg)
}
func (m *Message) RenderStatusBadRequest(arg ...string) *Message {
	return m.Render(RENDER_STATUS, http.StatusBadRequest, arg)
}
func (m *Message) RenderStatusUnauthorized(arg ...string) *Message {
	return m.Render(RENDER_STATUS, http.StatusUnauthorized, arg)
}
func (m *Message) RenderStatusForbidden(arg ...string) *Message {
	return m.Render(RENDER_STATUS, http.StatusForbidden, arg)
}
func (m *Message) RenderStatusNotFound(arg ...string) *Message {
	return m.Render(RENDER_STATUS, http.StatusNotFound, arg)
}
func (m *Message) RenderRedirect(arg ...Any) *Message {
	return m.Render(RENDER_REDIRECT, arg...)
}
func (m *Message) RenderDownload(arg ...Any) *Message {
	return m.Render(RENDER_DOWNLOAD, arg...)
}
func (m *Message) RenderResult(arg ...Any) *Message {
	return m.Render(RENDER_RESULT, arg...)
}
func (m *Message) RenderJson(arg ...Any) *Message {
	return m.Render(RENDER_JSON, kit.Format(kit.Dict(arg...)))
}

func (m *Message) IsCliUA() bool {
	if m.Option(MSG_USERUA) == "" || !strings.HasPrefix(m.Option(MSG_USERUA), "Mozilla") {
		return true
	}
	return false
}
func (m *Message) IsMobileUA() bool {
	return strings.Contains(m.Option(MSG_USERUA), "Mobile")
}
func (m *Message) PushSearch(arg ...Any) {
	data := kit.Dict(arg...)
	for i := 0; i < len(arg); i += 2 {
		switch k := arg[i].(type) {
		case string:
			if i+1 < len(arg) {
				data[k] = arg[i+1]
			}
		}
	}
	for _, k := range kit.Split(m.OptionFields()) {
		switch k {
		case TIME:
			m.Push(k, kit.Select(m.Time(), data[k]))
		case POD:
			m.Push(k, kit.Select("", data[k]))
		case CTX:
			m.Push(k, kit.Select(m.Prefix(), data[k]))
		case CMD:
			m.Push(k, kit.Select(m.CommandKey(), data[k]))
		default:
			m.Push(k, kit.Select("", data[k]))
		}
	}
}
func (m *Message) PushAction(arg ...Any) *Message {
	if len(m.meta[MSG_APPEND]) == 0 {
		return m
	}
	return m.Set(MSG_APPEND, ACTION).Tables(func(value Maps) { m.PushButton(arg...) })
}

func (m *Message) PushButton(arg ...Any) *Message {
	if !m.IsCliUA() {
		if m.FieldsIsDetail() {
			for i, k := range m.meta[KEY] {
				if k == ACTION {
					m.meta[VALUE][i] = Render(m, RENDER_BUTTON, arg...)
					return m
				}
			}
		} else if len(m.meta[ACTION]) >= m.Length() {
			m.meta[ACTION] = []string{}
		}
		m.Push(ACTION, Render(m, RENDER_BUTTON, arg...))
	}
	return m
}
func (m *Message) PushAnchor(arg ...string) {
	if !m.IsCliUA() {
		m.Push(LINK, Render(m, RENDER_ANCHOR, arg))
	}
}
func (m *Message) PushQRCode(key, src string) {
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_QRCODE, src))
	}
}
func (m *Message) PushImages(key, src string) {
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_IMAGES, src))
	}
}
func (m *Message) PushVideos(key, src string) {
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_VIDEOS, src))
	}
}
func (m *Message) PushIFrame(key, src string) {
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_IFRAME, src))
	}
}
func (m *Message) PushScript(arg ...string) {
	if !m.IsCliUA() {
		m.Push(SCRIPT, Render(m, RENDER_SCRIPT, arg))
	}
}
func (m *Message) PushDownload(key string, arg ...string) {
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_DOWNLOAD, arg))
	}
}

func (m *Message) EchoButton(arg ...Any) *Message {
	return m.Echo(Render(m, RENDER_BUTTON, arg...))
}
func (m *Message) EchoAnchor(arg ...string) *Message {
	return m.Echo(Render(m, RENDER_ANCHOR, arg))
}
func (m *Message) EchoQRCode(src string) *Message {
	return m.Echo(Render(m, RENDER_QRCODE, src))
}
func (m *Message) EchoImages(src string) *Message {
	return m.Echo(Render(m, RENDER_IMAGES, src))
}
func (m *Message) EchoVideos(src string) *Message {
	return m.Echo(Render(m, RENDER_VIDEOS, src))
}
func (m *Message) EchoIFrame(src string) *Message {
	return m.Echo(Render(m, RENDER_IFRAME, src))
}
func (m *Message) EchoScript(arg ...string) *Message {
	return m.Echo(Render(m, RENDER_SCRIPT, arg))
}
func (m *Message) EchoDownload(arg ...string) *Message {
	return m.Echo(Render(m, RENDER_DOWNLOAD, arg))
}
