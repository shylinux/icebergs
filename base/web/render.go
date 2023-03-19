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
	RESIZE  = "resize"
	OUTPUT  = "output"
	INPUT   = "input"
	VIEW    = "view"

	CODE_VIMER = "web.code.vimer"
	CODE_INNER = "web.code.inner"
	CODE_XTERM = "web.code.xterm"
	WIKI_WORD  = "web.wiki.word"
	CHAT_FAVOR = "web.chat.favor"
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
	m.W.Header().Add("Access-Control-Allow-Origin", "http://localhost:9020")
	switch cmd {
	case COOKIE: // value [name [path [expire]]]
		RenderCookie(m, arg[0], arg[1:]...)

	case STATUS, ice.RENDER_STATUS: // [code [text]]
		RenderStatus(m.W, kit.Int(kit.Select("200", arg, 0)), strings.Join(kit.Slice(arg, 1), " "))

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
		if cmd != "" && cmd != ice.RENDER_RAW {
			m.Echo(kit.Format(cmd, args...))
		}
		RenderType(m.W, nfs.JSON, "")
		m.DumpMeta(m.W)
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
	if m.IsCliUA() {
		return m.RenderDownload(path.Join(ice.USR_INTSHELL, kit.Select(ice.INDEX_SH, path.Join(file...))))
	}
	return m.RenderDownload(path.Join(ice.USR_VOLCANOS, kit.Select("page/"+ice.INDEX_HTML, path.Join(file...))))

	if repos == "" {
		repos = kit.Select(ice.VOLCANOS, ice.INTSHELL, m.IsCliUA())
	}
	p := func() string {
		defer mdb.RLock(m, "web.serve")()
		return path.Join(m.Conf(SERVE, kit.Keym(repos, nfs.PATH)), kit.Select(m.Conf(SERVE, kit.Keym(repos, INDEX)), path.Join(file...)))
	}
	return m.RenderDownload(p())
}
func RenderMain(m *ice.Message, pod, index string, arg ...ice.Any) *ice.Message {
	if script := m.Cmdx(Space(m, pod), nfs.CAT, kit.Select(ice.SRC_MAIN_JS, index)); script != "" {
		return m.Echo(kit.Renders(m.Cmdx(nfs.CAT, path.Join(ice.SRC_TEMPLATE, "web/main.html")), ice.Maps{nfs.VERSION: renderVersion(m), nfs.SCRIPT: script})).RenderResult()
	}
	return RenderIndex(m, ice.VOLCANOS)
}
func RenderCmd(m *ice.Message, cmd string, arg ...ice.Any) {
	RenderPodCmd(m, "", cmd, arg...)
}
func RenderCmds(m *ice.Message, list ...ice.Any) {
	m.Echo(kit.Renders(m.Cmdx(nfs.CAT, path.Join(ice.SRC_TEMPLATE, "web/cmd.html")), ice.Maps{nfs.VERSION: renderVersion(m), ice.LIST: kit.Format(list)})).RenderResult()
}
func RenderPodCmd(m *ice.Message, pod, cmd string, arg ...ice.Any) {
	msg := m.Cmd(Space(m, pod), ctx.COMMAND, kit.Select("web.wiki.word", cmd))
	list := kit.Format(kit.List(kit.Dict(msg.AppendSimple(mdb.NAME, mdb.HELP),
		ctx.INDEX, cmd, ctx.ARGS, kit.Simple(arg), ctx.DISPLAY, m.Option(ice.MSG_DISPLAY),
		mdb.LIST, kit.UnMarshal(msg.Append(mdb.LIST)), mdb.META, kit.UnMarshal(msg.Append(mdb.META)),
	)))
	m.Echo(kit.Renders(m.Cmdx(nfs.CAT, path.Join(ice.SRC_TEMPLATE, "web/cmd.html")), ice.Maps{nfs.VERSION: renderVersion(m), ice.LIST: list})).RenderResult()
}
func renderVersion(m *ice.Message) string {
	if strings.Contains(m.R.URL.RawQuery, "debug=true") {
		return kit.Format("?_v=%v&_t=%d", ice.Info.Make.Version, time.Now().Unix())
	}
	return ""
}
