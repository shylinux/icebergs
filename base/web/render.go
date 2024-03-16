package web

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/log"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const (
	STATUS   = "status"
	HEADER   = "header"
	COOKIE   = "cookie"
	REQUEST  = "request"
	RESPONSE = "response"
	CONTENT  = "content"
	TITLE    = "title"
	STYLE    = "style"
)

func renderMsg(m *ice.Message) {
	m.FormatsMeta(m.W,
		ice.MSG_USERIP, ice.MSG_USERUA, ice.MSG_METHOD, ice.MSG_REFERER, ice.MSG_DAEMON,
		ice.MSG_LANGUAGE, ice.MSG_THEME, ice.MSG_BG, ice.MSG_FG,
		ice.MSG_RIVER, ice.MSG_STORM, ice.MSG_INDEX, ice.MSG_FIELDS,
		ice.MSG_SOURCE, ice.MSG_TARGET,
		"task.id", "work.id", "space.timeout",
		ice.MSG_USERWEB0, ice.MSG_USERPOD0,
	)
}
func Render(m *ice.Message, cmd string, args ...ice.Any) bool {
	if cmd == ice.RENDER_VOID {
		return true
	}
	arg := kit.Simple(args...)
	kit.If(len(arg) == 0, func() { args = nil })
	if cmd != "" {
		if cmd != ice.RENDER_DOWNLOAD || !kit.HasPrefix(arg[0], nfs.SRC, nfs.USR, ice.SRC_TEMPLATE, ice.USR_INTSHELL, ice.USR_VOLCANOS) {
			if !(cmd == ice.RENDER_RESULT && len(args) == 0) {
				defer func() { m.Logs("Render", cmd, args) }()
			}
		}
	}
	switch cmd {
	case COOKIE: // value [name [path [expire]]]
		RenderCookie(m, arg[0], arg[1:]...)
	case STATUS, ice.RENDER_STATUS: // [code [text]]
		if m.IsCliUA() {
			RenderStatus(m.W, kit.Int(kit.Select("200", arg, 0)), kit.Select(m.Result(), strings.Join(kit.Slice(arg, 1), " ")))
		} else {
			m.W.WriteHeader(kit.Int(kit.Select("200", arg, 0)))
			renderMsg(m)
		}
	case ice.RENDER_REDIRECT: // url [arg...]
		http.Redirect(m.W, m.R, m.MergeLink(arg[0], arg[1:]), http.StatusTemporaryRedirect)
	case ice.RENDER_DOWNLOAD: // file [type [name]]
		if strings.HasPrefix(arg[0], HTTP) {
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
			kit.If(m.Result() == "" && m.Length() > 0, func() { m.TableEcho() })
			m.W.Write([]byte(m.Result()))
		}
	case ice.RENDER_JSON:
		RenderType(m.W, nfs.JSON, "")
		m.W.Write([]byte(arg[0]))
	default:
		kit.If(cmd != "" && cmd != ice.RENDER_RAW, func() { m.Echo(kit.Format(cmd, args...)) })
		RenderType(m.W, nfs.JSON, "")
		renderMsg(m)
	}
	m.Render(ice.RENDER_VOID)
	return true
}

func CookieName(url string) string { return ice.MSG_SESSID + "_" + kit.ParseURLMap(url)[tcp.PORT] }
func RenderCookie(m *ice.Message, value string, arg ...string) string { // name path expire
	http.SetCookie(m.W, &http.Cookie{Value: value, Name: kit.Select(CookieName(m.Option(ice.MSG_USERWEB)), arg, 0),
		Path: kit.Select(nfs.PS, arg, 1), Expires: time.Now().Add(kit.Duration(kit.Select(mdb.MONTH, arg, 2)))})
	return value
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
	RenderHeader(w, html.ContentType, mime)
}
func RenderOrigin(w http.ResponseWriter, origin string) {
	RenderHeader(w, "Access-Control-Allow-Origin", origin)
}
func RenderHeader(w http.ResponseWriter, key, value string) { w.Header().Set(key, value) }
func RenderStatus(w http.ResponseWriter, code int, text string) {
	w.WriteHeader(code)
	w.Write([]byte(text))
}
func RenderRedirect(m *ice.Message, arg ...ice.Any) { Render(m, ice.RENDER_REDIRECT, arg...) }
func RenderDownload(m *ice.Message, arg ...ice.Any) { Render(m, ice.RENDER_DOWNLOAD, arg...) }
func RenderResult(m *ice.Message, arg ...ice.Any)   { Render(m, ice.RENDER_RESULT, arg...) }
func RenderTemplate(m *ice.Message, file string, arg ...ice.Any) *ice.Message {
	return m.RenderResult(kit.Renders(kit.Format(m.Cmdx(nfs.CAT, path.Join(ice.SRC_TEMPLATE, WEB, file)), arg...), m))
}
func RenderMain(m *ice.Message) *ice.Message {
	if m.IsCliUA() {
		return m.RenderDownload(path.Join(ice.USR_INTSHELL, ice.INDEX_SH))
	}
	m.Options(nfs.SCRIPT, ice.SRC_MAIN_JS, nfs.VERSION, RenderVersion(m))
	return m.RenderResult(kit.Renders(m.Cmdx(nfs.CAT, ice.SRC_MAIN_HTML), m))
}
func RenderCmds(m *ice.Message, cmds ...ice.Any) {
	RenderMain(m.Options(ctx.CMDS, kit.Format(cmds)))
}
func RenderPodCmd(m *ice.Message, pod, cmd string, arg ...ice.Any) {
	if msg := m.Cmd(Space(m, pod), ctx.COMMAND, kit.Select(m.ShortKey(), cmd)); msg.Length() == 0 {
		RenderResult(m, kit.Format("not found command %s", cmd))
	} else {
		RenderCmds(m, kit.Dict(msg.AppendSimple(), ctx.ARGS, kit.Simple(arg), ctx.DISPLAY, m.Option(ice.MSG_DISPLAY)))
	}
}
func RenderCmd(m *ice.Message, cmd string, arg ...ice.Any) { RenderPodCmd(m, "", cmd, arg...) }

func RenderVersion(m *ice.Message) string {
	if ice.Info.Make.Hash == "" {
		return ""
	}
	ls := []string{ice.Info.Make.Version, ice.Info.Make.Forword, ice.Info.Make.Hash[:6]}
	if m.Option(log.DEBUG) == ice.TRUE || m.R != nil && strings.Contains(m.R.URL.RawQuery, "debug=true") {
		ls = append(ls, kit.Format("%d", time.Now().Unix()-kit.Time(ice.Info.Make.When)/int64(time.Second)))
	}
	return "?" + kit.JoinQuery(kit.Simple(kit.Dict("_v", strings.Join(ls, "-"), ice.POD, m.Option(ice.MSG_USERPOD)))...)
}

const (
	CHAT    = "chat"
	WORD    = "word"
	VIMER   = "vimer"
	XTERM   = "xterm"
	GRANT   = "grant"
	DESKTOP = "desktop"
	MESSAGE = "message"

	AUTOGEN = "autogen"
	BINPACK = "binpack"
	COMPILE = "compile"
	PUBLISH = "publish"
	VERSION = "version"
	UPGRADE = "upgrade"
	INSTALL = "install"

	CODE_GIT_SERVICE  = "web.code.git.service"
	CODE_GIT_SEARCH   = "web.code.git.search"
	CODE_GIT_STATUS   = "web.code.git.status"
	CODE_GIT_REPOS    = "web.code.git.repos"
	CODE_AUTOGEN      = "web.code.autogen"
	CODE_COMPILE      = "web.code.compile"
	CODE_PUBLISH      = "web.code.publish"
	CODE_UPGRADE      = "web.code.upgrade"
	CODE_VIMER        = "web.code.vimer"
	CODE_INNER        = "web.code.inner"
	CODE_XTERM        = "web.code.xterm"
	WIKI_FEEL         = "web.wiki.feel"
	WIKI_DRAW         = "web.wiki.draw"
	WIKI_WORD         = "web.wiki.word"
	WIKI_PORTAL       = "web.wiki.portal"
	CHAT_OAUTH_CLIENT = "web.chat.oauth.client"
	CHAT_MESSAGE      = "web.chat.message"
	CHAT_HEADER       = "web.chat.header"
	CHAT_IFRAME       = "web.chat.iframe"
	CHAT_FAVOR        = "web.chat.favor"
	CHAT_FLOWS        = "web.chat.flows"
	CHAT_GRANT        = "web.chat.grant"
	CHAT_POD          = "web.chat.pod"
	CHAT_CMD          = "web.chat.cmd"
	TEAM_PLAN         = "web.team.plan"
)

func MessageInsertJSON(m *ice.Message, zone, name, text string, arg ...string) {
	MessageInsert(m, zone, kit.Simple(mdb.TYPE, html.TEXT, mdb.NAME, kit.Select(m.CommandKey(), name),
		mdb.TEXT, text, ctx.DISPLAY, html.PLUGIN_JSON, arg)...)
}
func MessageInsert(m *ice.Message, zone string, arg ...string) {
	m.Cmd(MESSAGE, mdb.INSERT, zone, tcp.DIRECT, tcp.RECV, arg)
}
