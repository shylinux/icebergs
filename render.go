package ice

import (
	"net/http"
	"path"
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
		if !strings.HasPrefix(p, HTTP) {
			p = m.MergeURL2(p)
		}
		return kit.Format(`<a href="%s" target="_blank">%s</a>`, p, arg[0])

	case RENDER_BUTTON: // name...
		list := []string{}
		for _, k := range kit.Split(kit.Join(arg)) {
			list = append(list, kit.Format(`<input type="button" name="%s" value="%s">`, k,
				kit.Select(k, kit.Value(m._cmd.Meta, kit.Keys("_trans", k)), m.Option(MSG_LANGUAGE) != "en")))
		}
		return kit.Join(list, SP)

	case RENDER_IMAGES: // src [height]
		return kit.Format(`<img src="%s" height=%s>`, arg[0], kit.Select("120", arg, 1))

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
	m.Optionv(MSG_OUTPUT, cmd)
	m.Optionv(MSG_ARGS, args)

	switch cmd {
	case RENDER_TEMPLATE: // text [data]
		if len(args) == 1 {
			args = append(args, m)
		}
		if res, err := kit.Render(args[0].(string), args[1]); m.Assert(err) {
			m.Echo(string(res))
		}
	}
	return m
}
func (m *Message) RenderVoid() *Message {
	return m.Render(RENDER_VOID)
}
func (m *Message) RenderJson(args ...Any) *Message { // [key value]...
	return m.Render(RENDER_JSON, kit.Format(kit.Dict(args...)))
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
func (m *Message) RenderResult(args ...Any) *Message { // [fmt arg...]
	return m.Render(RENDER_RESULT, args...)
}
func (m *Message) RenderTemplate(args ...Any) *Message {
	return m.Render(RENDER_TEMPLATE, args...)
}
func (m *Message) RenderRedirect(args ...Any) *Message {
	return m.Render(RENDER_REDIRECT, args...)
}
func (m *Message) RenderDownload(args ...Any) *Message {
	return m.Render(RENDER_DOWNLOAD, args...)
}
func (m *Message) RenderWebsite(pod string, dir string, arg ...string) *Message {
	m.Echo(m.Cmdx(m.Space(pod), WEBSITE, "parse", dir, arg))
	return m.RenderResult()
}
func (m *Message) RenderIndex(serve, repos string, file ...string) *Message {
	return m.RenderDownload(path.Join(m.Conf(serve, kit.Keym(repos, "path")), kit.Select(m.Conf(serve, kit.Keym(repos, INDEX)), path.Join(file...))))
}
func (m *Message) RenderCmd(index string, args ...Any) {
	list := index
	if index != "" {
		msg := m.Cmd(COMMAND, index)
		list = kit.Format(kit.List(kit.Dict(msg.AppendSimple(NAME, HELP),
			INDEX, index, ARGS, kit.Simple(args), DISPLAY, m.Option(MSG_DISPLAY),
			INPUTS, kit.UnMarshal(msg.Append(LIST)), FEATURE, kit.UnMarshal(msg.Append(META)),
		)))
	}
	m.Echo(kit.Format(Info.cans, list))
	m.RenderResult()
}

func (m *Message) IsMobileUA() bool {
	return strings.Contains(m.Option(MSG_USERUA), "Mobile")
}
func (m *Message) IsCliUA() bool {
	if m.Option(MSG_USERUA) == "" || !strings.HasPrefix(m.Option(MSG_USERUA), "Mozilla/5.0") {
		return true
	}
	return false
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
func (m *Message) PushScript(arg ...string) { // [type] text...
	if !m.IsCliUA() {
		m.Push(SCRIPT, Render(m, RENDER_SCRIPT, arg))
	}
}
func (m *Message) PushQRCode(key string, src string, arg ...string) { // key src [height]
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_QRCODE, src, arg))
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
func (m *Message) PushDownload(key string, arg ...Any) { // [name] file
	if !m.IsCliUA() {
		m.Push(key, Render(m, RENDER_DOWNLOAD, arg...))
	}
}
func (m *Message) PushAction(list ...Any) *Message {
	if len(m.meta[MSG_APPEND]) == 0 {
		return m
	}
	return m.Set(MSG_APPEND, ACTION).Tables(func(value Maps) { m.PushButton(list...) })
}
func (m *Message) PushSearch(args ...Any) {
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
		m.Tables(func(value Maps) { m.Push(POD, m.Option(MSG_USERPOD)) })
	}

	m.Cmd(SPACE, OptionFields("type,name")).Tables(func(value Maps) {
		switch value[TYPE] {
		case "server", "worker":
			if value[NAME] == Info.HostName {
				break
			}
			m.Cmd(SPACE, value[NAME], m.Prefix(cmd), arg).Table(func(index int, val Maps, head []string) {
				val[POD] = kit.Keys(value[NAME], val[POD])
				m.Push("", val, head)
			})
		}
	})
}

func (m *Message) EchoAnchor(arg ...Any) *Message { // [name] link
	return m.Echo(Render(m, RENDER_ANCHOR, arg...))
}
func (m *Message) EchoButton(arg ...Any) *Message { // name...
	return m.Echo(Render(m, RENDER_BUTTON, arg...))
}
func (m *Message) EchoScript(arg ...string) *Message { // [type] text...
	return m.Echo(Render(m, RENDER_SCRIPT, arg))
}
func (m *Message) EchoQRCode(src string, arg ...string) *Message { // src [height]
	return m.Echo(Render(m, RENDER_QRCODE, src, arg))
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
func (m *Message) EchoDownload(arg ...Any) *Message { // [name] file
	return m.Echo(Render(m, RENDER_DOWNLOAD, arg...))
}

func (m *Message) DisplayBase(file string, arg ...Any) *Message {
	if !strings.Contains(file, PT) {
		file += ".js"
	}
	m.Option(MSG_DISPLAY, kit.MergeURL(DisplayBase(file)[DISPLAY], arg...))
	return m
}
func (m *Message) DisplayStory(file string, arg ...Any) *Message { // /plugin/story/file
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PLUGIN_STORY, file)
	}
	return m.DisplayBase(file, arg...)
}
func (m *Message) DisplayLocal(file string, arg ...Any) *Message { // /plugin/local/file
	if file == "" {
		file = path.Join(kit.PathName(2), kit.Keys(kit.FileName(2), JS))
	}
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PLUGIN_LOCAL, file)
	}
	return m.DisplayBase(file, arg...)
}
func (m *Message) Display(file string, arg ...Any) *Message { // repos local file
	m.Option(MSG_DISPLAY, kit.MergeURL(displayRequire(2, file)[DISPLAY], arg...))
	return m
}
func (m *Message) DisplayStoryJSON(arg ...Any) *Message { // /plugin/story/json.js
	return m.DisplayStory("json", arg...)
}

func DisplayBase(file string, arg ...string) Maps {
	return Maps{DISPLAY: file, STYLE: kit.Join(arg, SP)}
}
func DisplayStory(file string, arg ...string) Maps { // /plugin/story/file
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PLUGIN_STORY, file)
	}
	return DisplayBase(file, arg...)
}
func DisplayLocal(file string, arg ...string) Maps { // /plugin/local/file
	if file == "" {
		file = path.Join(kit.PathName(2), kit.Keys(kit.FileName(2), JS))
	}
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PLUGIN_LOCAL, file)
	}
	return DisplayBase(file, arg...)
}
func Display(file string, arg ...string) Maps { // repos local file
	return displayRequire(2, file, arg...)
}
func displayRequire(n int, file string, arg ...string) Maps {
	if file == "" {
		file = kit.Keys(kit.FileName(n+1), JS)
	}
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PS, path.Join(path.Dir(FileRequire(n+2)), file))
	}
	return DisplayBase(file, arg...)
}
