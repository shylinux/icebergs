package web

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

const (
	COOKIE = "cookie"
	STATUS = "status"
)
const (
	WEBSITE = "website"

	CODE_VIMER = "web.code.vimer"
	CODE_INNER = "web.code.inner"
	WIKI_WORD  = "web.wiki.word"
)

func Render(m *ice.Message, cmd string, args ...ice.Any) bool {
	if cmd == ice.RENDER_VOID {
		return true
	}
	arg := kit.Simple(args...)
	if len(arg) == 0 {
		args = nil
	}
	if cmd != "" && cmd != ice.RENDER_DOWNLOAD {
		defer func() { m.Logs("Render", cmd, args) }()
	}
	switch cmd {
	case COOKIE: // value [name [path [expire]]]
		RenderCookie(m, arg[0], arg[1:]...)

	case STATUS, ice.RENDER_STATUS: // [code [text]]
		RenderStatus(m.W, kit.Int(kit.Select("200", arg, 0)), kit.Select("", arg, 1))

	case ice.RENDER_REDIRECT: // url [arg...]
		http.Redirect(m.W, m.R, kit.MergeURL(arg[0], arg[1:]), http.StatusTemporaryRedirect)

	case ice.RENDER_DOWNLOAD: // file [type [name]]
		if strings.HasPrefix(arg[0], ice.HTTP) {
			RenderRedirect(m, arg[0])
			break
		}
		RenderType(m.W, arg[0], kit.Select("", arg, 1))
		RenderHeader(m.W, "Content-Disposition", fmt.Sprintf("filename=%s", kit.Select(path.Base(kit.Select(arg[0], m.Option("filename"))), arg, 2)))
		if _, e := nfs.DiskFile.StatFile(arg[0]); e == nil {
			http.ServeFile(m.W, m.R, kit.Path(arg[0]))
		} else if f, e := nfs.PackFile.OpenFile(arg[0]); e == nil {
			defer f.Close()
			io.Copy(m.W, f)
		}

	case ice.RENDER_RESULT:
		if len(arg) > 0 { // [str [arg...]]
			m.W.Write([]byte(kit.Format(arg[0], args[1:]...)))
		} else {
			if m.Result() == "" && m.Length() > 0 {
				m.Table()
			}
			m.W.Write([]byte(m.Result()))
		}

	case ice.RENDER_JSON:
		RenderType(m.W, nfs.JSON, "")
		m.W.Write([]byte(arg[0]))

	default:
		for _, k := range kit.Simple(m.Optionv("option"), m.Optionv("_option")) {
			if m.Option(k) == "" {
				m.Set(k)
			}
		}
		for _, k := range []string{"sessid", "cmds", "fields", "_option", "_handle", "_output"} {
			m.Set(k)
		}
		if cmd != "" && cmd != ice.RENDER_RAW {
			m.Echo(kit.Format(cmd, args...))
		}
		RenderType(m.W, nfs.JSON, "")
		fmt.Fprint(m.W, m.FormatMeta())
	}
	return true
}

func RenderType(w http.ResponseWriter, name, mime string) {
	if mime == "" {
		switch kit.Ext(name) {
		case "", nfs.HTML:
			mime = "text/html"
		case nfs.CSS:
			mime = "text/css; charset=utf-8"
		default:
			mime = "application/" + kit.Ext(name)
		}
	}
	RenderHeader(w, ContentType, mime)
}
func RenderHeader(w http.ResponseWriter, key, value string) {
	w.Header().Set(key, value)
}
func RenderCookie(m *ice.Message, value string, arg ...string) { // name path expire
	expire := time.Now().Add(kit.Duration(kit.Select(m.Conf(aaa.SESS, kit.Keym(mdb.EXPIRE)), arg, 2)))
	http.SetCookie(m.W, &http.Cookie{Value: value,
		Name: kit.Select(CookieName(m.Option(ice.MSG_USERWEB)), arg, 0), Path: kit.Select(ice.PS, arg, 1), Expires: expire})
}
func RenderStatus(w http.ResponseWriter, code int, text string) {
	w.WriteHeader(code)
	w.Write([]byte(text))
}
func RenderRefresh(m *ice.Message, arg ...string) { // url text delay
	Render(m, ice.RENDER_RESULT, kit.Format(`
<html>
<head>
	<meta http-equiv="refresh" content="%s; url='%s'">
</head>
<body>
	%s
</body>
</html>
`, kit.Select("3", arg, 2), kit.Select(m.Option(ice.MSG_USERWEB), arg, 0), kit.Select("loading...", arg, 1)))
	m.Render(ice.RENDER_VOID)
}
func RenderRedirect(m *ice.Message, arg ...ice.Any) {
	Render(m, ice.RENDER_REDIRECT, arg...)
	m.Render(ice.RENDER_VOID)
}
func RenderDownload(m *ice.Message, arg ...ice.Any) {
	Render(m, ice.RENDER_DOWNLOAD, arg...)
	m.Render(ice.RENDER_VOID)
}
func RenderResult(m *ice.Message, arg ...ice.Any) {
	Render(m, ice.RENDER_RESULT, arg...)
	m.Render(ice.RENDER_VOID)
}

func CookieName(url string) string {
	return ice.MSG_SESSID + "_" + kit.ReplaceAll(kit.ParseURLMap(url)[tcp.PORT], ".", "_", ":", "_")
	return ice.MSG_SESSID + "_" + kit.ReplaceAll(kit.ParseURLMap(url)[tcp.HOST], ".", "_", ":", "_")
}

func RenderIndex(m *ice.Message, repos string, file ...string) *ice.Message {
	if repos == "" {
		repos = kit.Select(ice.VOLCANOS, ice.INTSHELL, m.IsCliUA())
	}
	return m.RenderDownload(path.Join(m.Conf(SERVE, kit.Keym(repos, nfs.PATH)), kit.Select(m.Conf(SERVE, kit.Keym(repos, INDEX)), path.Join(file...))))
}
func RenderMain(m *ice.Message, pod, index string, arg ...ice.Any) *ice.Message {
	if script := m.Cmdx(Space(m, pod), nfs.CAT, kit.Select(ice.SRC_MAIN_JS, index)); script != "" {
		return m.Echo(kit.Renders(_main_template, ice.Maps{"version": renderVersion(m), "script": script})).RenderResult()
	}
	return RenderIndex(m, ice.VOLCANOS)
}
func RenderCmd(m *ice.Message, cmd string, arg ...ice.Any) {
	RenderPodCmd(m, "", cmd, arg...)
}
func RenderPodCmd(m *ice.Message, pod, cmd string, arg ...ice.Any) {
	msg := m.Cmd(Space(m, pod), ctx.COMMAND, kit.Select("web.wiki.word", cmd))
	list := kit.Format(kit.List(kit.Dict(msg.AppendSimple(mdb.NAME, mdb.HELP),
		ctx.INDEX, cmd, ctx.ARGS, kit.Simple(arg), ctx.DISPLAY, m.Option(ice.MSG_DISPLAY),
		mdb.LIST, kit.UnMarshal(msg.Append(mdb.LIST)), mdb.META, kit.UnMarshal(msg.Append(mdb.META)),
	)))
	m.Echo(kit.Renders(_cmd_template, ice.Maps{
		"version": renderVersion(m), "list": list,
	})).RenderResult()
}
func renderVersion(m *ice.Message) string {
	if strings.Contains(m.R.URL.RawQuery, "debug=true") {
		return kit.Format("?_v=%v&_t=%d", ice.Info.Make.Version, time.Now().Unix())
	}
	return ""
}

var _main_template = `<!DOCTYPE html>
<head>
	<meta name="viewport" content="width=device-width,initial-scale=0.8,maximum-scale=0.8,user-scalable=no">
	<meta charset="utf-8"><title>volcanos</title>
	<link href="/favicon.ico" rel="shortcut icon" type="image/ico">
	<link href="/page/cache.css{{.version}}" rel="stylesheet">
	<link href="/page/index.css{{.version}}" rel="stylesheet">
</head>
<body>
	<script>_version = "{{.version}}"</script>
	<script src="/proto.js{{.version}}"></script>
	<script src="/page/cache.js{{.version}}"></script>
	<script>{{.script}}</script>
</body>
`

var _cmd_template = `<!DOCTYPE html>
<head>
	<meta name="viewport" content="width=device-width,initial-scale=0.8,maximum-scale=0.8,user-scalable=no">
	<meta charset="utf-8"><title>volcanos</title>
	<link href="/page/can.css{{.version}}" rel="stylesheet">
</head>
<body>
	<script>_version = "{{.version}}"</script>
	<script src="/page/can.js{{.version}}"></script><script>Volcanos({{.list}})</script>
</body>
`
