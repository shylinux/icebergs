package ice

import (
	"net/http"
	"path"
	"strings"

	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func AddRender(key string, render func(*Message, ...Any) string) { Info.render[key] = render }
func RenderAction(key ...string) Actions {
	return Actions{CTX_INIT: {Hand: func(m *Message, arg ...string) {
		cmd := m.CommandKey()
		kit.For(key, func(key string) {
			AddRender(key, func(m *Message, arg ...Any) string { return m.Cmd(cmd, key, arg).Result() })
		})
	}}}
}
func Render(m *Message, cmd string, args ...Any) string {
	if render, ok := Info.render[cmd]; ok {
		return render(m, args...)
	}
	trans := kit.Value(m._cmd.Meta, CTX_TRANS)
	switch arg := kit.Simple(args...); cmd {
	case RENDER_BUTTON:
		list := []string{}
		for _, k := range args {
			switch k := k.(type) {
			case []string:
				kit.For(k, func(k string) { list = append(list, Render(m, RENDER_BUTTON, k)) })
			case string:
				if k == "" {
					break
				}
				if strings.HasPrefix(k, "<input") {
					list = append(list, k)
					break
				}
				kit.For(kit.Split(k), func(k string) {
					list = append(list, kit.Format(`<input type="button" name="%s" value="%s">`, k, kit.Select(k, kit.Value(trans, k), !m.IsEnglish())))
				})
			case Map, Maps:
				kit.For(k, func(k, v string) {
					list = append(list, kit.Format(`<input type="button" name="%s" value="%s">`, k, kit.Select(v, k, m.IsEnglish())))
				})
			default:
				list = append(list, Render(m, RENDER_BUTTON, kit.LowerCapital(kit.Format(k))))
			}
		}
		return strings.Join(list, "")
	case RENDER_ANCHOR:
		return kit.Format(`<a href="%s" target="_blank">%s</a>`, kit.Select(arg[0], arg, 1), arg[0])
	case RENDER_SCRIPT:
		return kit.Format(`<code>%s</code>`, arg[0])
	case RENDER_IMAGES:
		if len(arg) > 1 {
			return kit.Format(`<img src="%s" height="%s">`, arg[0], arg[1])
		}
		return kit.Format(`<img src="%s">`, arg[0])
	case RENDER_VIDEOS:
		return kit.Format(`<video src="%s" controls autoplay>`, arg[0])
	case RENDER_AUDIOS:
		return kit.Format(`<audio src="%s" controls autoplay>`, arg[0])
	case RENDER_IFRAME:
		return kit.Format(`<iframe src="%s"></iframe>`, arg[0])
	default:
		if len(arg) == 1 {
			return kit.Format(`<%s>%s</%s>`, cmd, arg[0], cmd)
		}
		return kit.Format(`<%s style="%s">%s</%s>`, cmd, kit.JoinKV(DF, ";", arg[1:]...), arg[0], cmd)
	}
}

func (m *Message) Render(cmd string, arg ...Any) *Message {
	switch cmd {
	case RENDER_TEMPLATE:
		kit.If(len(arg) == 1, func() { arg = append(arg, m) })
		if res, err := kit.Render(arg[0].(string), arg[1]); m.Assert(err) {
			m.Echo(string(res))
		}
		return m
	}
	return m.Options(MSG_OUTPUT, cmd, MSG_ARGS, arg)
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
func (m *Message) RenderTemplate(arg ...Any) *Message {
	return m.Render(RENDER_TEMPLATE, arg...)
}
func (m *Message) RenderResult(arg ...Any) *Message {
	return m.Render(RENDER_RESULT, arg...)
}
func (m *Message) RenderJson(arg ...Any) *Message {
	return m.Render(RENDER_JSON, kit.Format(kit.Dict(arg...)))
}
func (m *Message) RenderVoid(arg ...Any) *Message {
	return m.Render(RENDER_VOID, arg...)
}
func (m *Message) IsDebug() bool {
	return m.Option(MSG_DEBUG) == TRUE
}
func (m *Message) IsCliUA() bool {
	return m.Option(MSG_USERUA) == "" || !strings.HasPrefix(m.Option(MSG_USERUA), html.Mozilla)
}
func (m *Message) IsWeixinUA() bool {
	return strings.Contains(m.Option(MSG_USERUA), html.MicroMessenger)
}
func (m *Message) IsMobileUA() bool {
	return strings.Contains(m.Option(MSG_USERUA), html.Mobile)
}
func (m *Message) IsChromeUA() bool {
	return strings.Contains(m.Option(MSG_USERUA), html.Chrome)
}
func (m *Message) IsMetaKey() bool {
	return m.Option("metaKey") == TRUE
}
func (m *Message) IsGetMethod() bool {
	return m.Option(MSG_METHOD) == http.MethodGet
}
func (m *Message) IsLocalhost() bool {
	return strings.HasPrefix(m.Option(MSG_USERWEB), "http://localhost:9020")
}
func (m *Message) PushSearch(arg ...Any) {
	data := kit.Dict(arg...)
	kit.For(arg, func(k, v Any) {
		if k, ok := k.(string); ok {
			data[k] = v
		}
	})
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
	if len(m.value(MSG_APPEND)) == 0 {
		return m
	}
	m.Set(MSG_APPEND, ACTION)
	return m.Table(func(value Maps) { m.PushButton(arg...) })
}
func (m *Message) PushButton(arg ...Any) *Message {
	if !m.IsCliUA() {
		if m.FieldsIsDetail() {
			for i, k := range m.value(KEY) {
				if k == ACTION {
					m.index(VALUE, i, Render(m, RENDER_BUTTON, arg...))
					return m
				}
			}
		} else if len(m.value(ACTION)) >= m.Length() {
			m.delete(ACTION)
		}
		m.Push(ACTION, Render(m, RENDER_BUTTON, arg...))
	}
	return m
}
func (m *Message) PushAnchor(arg ...string) {
	kit.If(!m.IsCliUA(), func() { m.Push(LINK, Render(m, RENDER_ANCHOR, arg)) })
}
func (m *Message) PushQRCode(key, src string) *Message {
	kit.If(!m.IsCliUA(), func() { m.Push(key, Render(m, RENDER_QRCODE, src)) })
	return m
}
func (m *Message) PushScript(arg ...string) {
	kit.If(!m.IsCliUA(), func() { m.Push(SCRIPT, Render(m, RENDER_SCRIPT, arg)) })
}
func (m *Message) PushImages(key, src string, arg ...string) {
	kit.If(!m.IsCliUA(), func() { m.Push(key, Render(m, RENDER_IMAGES, src, arg)) })
}
func (m *Message) PushVideos(key, src string) {
	kit.If(!m.IsCliUA(), func() { m.Push(key, Render(m, RENDER_VIDEOS, src)) })
}
func (m *Message) PushAudios(key, src string) {
	kit.If(!m.IsCliUA(), func() { m.Push(key, Render(m, RENDER_AUDIOS, src)) })
}
func (m *Message) PushIFrame(key, src string) {
	kit.If(!m.IsCliUA(), func() { m.Push(key, Render(m, RENDER_IFRAME, src)) })
}
func (m *Message) PushDownload(key string, arg ...string) *Message {
	kit.If(!m.IsCliUA(), func() { m.Push(key, Render(m, RENDER_DOWNLOAD, arg)) })
	return m
}

func (m *Message) EchoInfoButton(info string, arg ...Any) *Message {
	kit.If(info == "", func() { info = Info.Template(m, m.ActionKey()+".html") })
	kit.If(len(arg) == 0, func() { arg = append(arg, m.ActionKey()) })
	m.Display("/plugin/table.js", "style", "form")
	return m.Echo(html.Format("div", info, "class", "info", "style", kit.JoinCSS())).EchoButton(arg...).Echo(NL).Action(arg...)
}
func (m *Message) EchoButton(arg ...Any) *Message {
	if len(arg) == 0 || len(arg) == 1 && arg[0] == nil {
		return m
	}
	return m.Echo(Render(m, RENDER_BUTTON, arg...))
}
func (m *Message) EchoAnchor(arg ...string) *Message { return m.Echo(Render(m, RENDER_ANCHOR, arg)) }
func (m *Message) EchoQRCode(src string) *Message    { return m.Echo(Render(m, RENDER_QRCODE, src)) }
func (m *Message) EchoScript(arg ...string) *Message { return m.Echo(Render(m, RENDER_SCRIPT, arg)) }
func (m *Message) EchoImages(src string) *Message    { return m.Echo(Render(m, RENDER_IMAGES, src)) }
func (m *Message) EchoVideos(src string) *Message    { return m.Echo(Render(m, RENDER_VIDEOS, src)) }
func (m *Message) EchoAudios(src string) *Message    { return m.Echo(Render(m, RENDER_AUDIOS, src)) }
func (m *Message) EchoIFrame(src string) *Message {
	kit.If(!m.IsCliUA(), func() {
		kit.If(src, func() { m.Echo(Render(m, RENDER_IFRAME, src)) })
	})
	return m
}
func (m *Message) EchoDownload(arg ...string) *Message {
	return m.Echo(Render(m, RENDER_DOWNLOAD, arg))
}
func (m *Message) EchoFields(cmd string, arg ...string) *Message {
	return m.Echo(`<fieldset class="story" data-index="%s" data-args=%q>
<legend></legend>
<form class="option"></form>
<div class="action"></div>
<div class="output"></div>
<div class="status"></div>
</fieldset>
`, cmd, kit.Join(arg))
}
func (m *Message) Display(file string, arg ...Any) *Message {
	m.Option(MSG_DISPLAY, kit.MergeURL(kit.ExtChange(m.resource(file), JS), arg...))
	return m
}
func (m *Message) Resource(file string, arg ...string) string {
	if len(arg) > 0 && arg[0] != "" {
		if strings.HasPrefix(file, HTTP) {
			return file
		} else if strings.HasPrefix(file, PS) {
			return arg[0] + file
		} else if kit.HasPrefix(file, "src", "usr") {
			return arg[0] + "/require/" + file
		}
	}
	return m.resource(file)
}
func (m *Message) resource(file string) string {
	p := kit.FileLines(3)
	kit.If(file == "", func() { p = kit.ExtChange(p, JS) }, func() {
		if kit.HasPrefix(file, PS, HTTP) {
			p = file
		} else {
			p = path.Join(path.Dir(p), file)
		}
	})
	return m.FileURI(p)
}
