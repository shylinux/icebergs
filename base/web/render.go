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
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

const (
	COOKIE = "cookie"
	STATUS = "status"
)

func Render(msg *ice.Message, cmd string, args ...ice.Any) {
	if cmd != "" {
		defer func() { msg.Logs(mdb.EXPORT, cmd, args) }()
	}

	switch arg := kit.Simple(args...); cmd {
	case COOKIE: // value [name [path [expire]]]
		RenderCookie(msg, arg[0], arg[1:]...)

	case STATUS, ice.RENDER_STATUS: // [code [text]]
		RenderStatus(msg.W, kit.Int(kit.Select("200", arg, 0)), kit.Select("", arg, 1))

	case ice.RENDER_REDIRECT: // url [arg...]
		http.Redirect(msg.W, msg.R, kit.MergeURL(arg[0], arg[1:]), http.StatusTemporaryRedirect)

	case ice.RENDER_DOWNLOAD: // file [type [name]]
		if strings.HasPrefix(arg[0], ice.HTTP) {
			RenderRedirect(msg, arg[0])
			break
		}
		RenderType(msg.W, arg[0], kit.Select("", arg, 1))
		RenderHeader(msg.W, "Content-Disposition", fmt.Sprintf("filename=%s", kit.Select(path.Base(kit.Select(arg[0], msg.Option("filename"))), arg, 2)))
		if _, e := nfs.DiskFile.StatFile(arg[0]); e == nil {
			http.ServeFile(msg.W, msg.R, kit.Path(arg[0]))
		} else if f, e := nfs.PackFile.OpenFile(arg[0]); e == nil {
			defer f.Close()
			io.Copy(msg.W, f)
		}

	case ice.RENDER_RESULT:
		if len(arg) > 0 { // [str [arg...]]
			msg.W.Write([]byte(kit.Format(arg[0], args[1:]...)))
		} else {
			args = append(args, nfs.SIZE, len(msg.Result()))
			msg.W.Write([]byte(msg.Result()))
		}

	case ice.RENDER_JSON:
		RenderType(msg.W, nfs.JSON, "")
		msg.W.Write([]byte(arg[0]))

	case ice.RENDER_VOID:
		// no output

	default:
		for _, k := range []string{
			"_option", "_handle", "_output",
			"cmds", "fields", "sessid",
		} {
			msg.Set(k)
		}

		if cmd != "" && cmd != ice.RENDER_RAW { // [str [arg...]]
			msg.Echo(kit.Format(cmd, args...))
		}
		RenderType(msg.W, nfs.JSON, "")
		fmt.Fprint(msg.W, msg.FormatMeta())
	}
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
func RenderCookie(msg *ice.Message, value string, arg ...string) { // name path expire
	expire := time.Now().Add(kit.Duration(kit.Select(msg.Conf(aaa.SESS, kit.Keym(mdb.EXPIRE)), arg, 2)))
	http.SetCookie(msg.W, &http.Cookie{Value: value,
		Name: kit.Select(CookieName(msg.Option(ice.MSG_USERWEB)), arg, 0), Path: kit.Select(ice.PS, arg, 1), Expires: expire})
}
func RenderStatus(w http.ResponseWriter, code int, text string) {
	w.WriteHeader(code)
	w.Write([]byte(text))
}
func RenderRefresh(msg *ice.Message, arg ...string) { // url text delay
	Render(msg, ice.RENDER_RESULT, kit.Format(`
<html>
<head>
	<meta http-equiv="refresh" content="%s; url='%s'">
</head>
<body>
	%s
</body>
</html>
`, kit.Select("3", arg, 2), kit.Select(msg.Option(ice.MSG_USERWEB), arg, 0), kit.Select("loading...", arg, 1)))
	msg.Render(ice.RENDER_VOID)
}
func RenderRedirect(msg *ice.Message, arg ...ice.Any) {
	Render(msg, ice.RENDER_REDIRECT, arg...)
	msg.Render(ice.RENDER_VOID)
}
func RenderDownload(msg *ice.Message, arg ...ice.Any) {
	Render(msg, ice.RENDER_DOWNLOAD, arg...)
	msg.Render(ice.RENDER_VOID)
}
func RenderResult(msg *ice.Message, arg ...ice.Any) {
	Render(msg, ice.RENDER_RESULT, arg...)
	msg.Render(ice.RENDER_VOID)
}

func CookieName(url string) string {
	return ice.MSG_SESSID + "_" + kit.ReplaceAll(kit.ParseURLMap(url)[tcp.HOST], ".", "_", ":", "_")
	return ice.MSG_SESSID + "_" + kit.ParseURLMap(url)[tcp.PORT]
}

func RenderIndex(m *ice.Message, repos string, file ...string) *ice.Message {
	if repos == "" {
		repos = kit.Select(ice.VOLCANOS, ice.INTSHELL, m.IsCliUA())
	}
	return m.RenderDownload(path.Join(m.Conf(SERVE, kit.Keym(repos, nfs.PATH)), kit.Select(m.Conf(SERVE, kit.Keym(repos, INDEX)), path.Join(file...))))
}
func RenderWebsite(m *ice.Message, pod string, dir string, arg ...string) *ice.Message {
	return m.Echo(m.Cmdx(Space(m, pod), "web.chat.website", lex.PARSE, dir, arg)).RenderResult()
}
func RenderCmd(m *ice.Message, index string, args ...ice.Any) {
	if index == "" {
		return
	}
	list := index
	if index != "" {
		msg := m.Cmd(ctx.COMMAND, index)
		list = kit.Format(kit.List(kit.Dict(msg.AppendSimple(mdb.NAME, mdb.HELP),
			ctx.INDEX, index, ctx.ARGS, kit.Simple(args), ctx.DISPLAY, m.Option(ice.MSG_DISPLAY),
			mdb.LIST, kit.UnMarshal(msg.Append(mdb.LIST)), mdb.META, kit.UnMarshal(msg.Append(mdb.META)),
		)))
	}
	m.Echo(kit.Format(_cans, list))
	m.RenderResult()
}

var _cans = `<!DOCTYPE html>
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=0.8,maximum-scale=0.8,user-scalable=no">
	<link rel="stylesheet" type="text/css" href="/page/can.css">
</head>
<body>
	<script src="/page/can.js"></script><script>Volcanos(%s)</script>
</body>
`
