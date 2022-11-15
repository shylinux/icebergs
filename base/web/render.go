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

func Render(m *ice.Message, cmd string, args ...ice.Any) bool {
	if cmd != "" {
		defer func() { m.Logs(mdb.EXPORT, cmd, args) }()
	}

	switch arg := kit.Simple(args...); cmd {
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
			args = append(args, nfs.SIZE, len(m.Result()))
			m.W.Write([]byte(m.Result()))
		}

	case ice.RENDER_JSON:
		RenderType(m.W, nfs.JSON, "")
		m.W.Write([]byte(arg[0]))

	case ice.RENDER_VOID:
		// no output

	default:
		for _, k := range kit.Simple(m.Optionv("option"), m.Optionv("_option")) {
			if m.Option(k) == "" {
				m.Set(k)
			}
		}
		for _, k := range []string{"sessid", "cmds", "fields", "_option", "_handle", "_output"} {
			m.Set(k)
		}

		if cmd != "" && cmd != ice.RENDER_RAW { // [str [arg...]]
			m.Echo(kit.Format(cmd, args...))
		}
		RenderType(m.W, nfs.JSON, "")
		fmt.Fprint(m.W, m.FormatsMeta())
		// fmt.Fprint(m.W, m.FormatMeta())
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
	m.Echo(kit.Format(_cmd_template, list)).RenderResult()
}
func RenderMain(m *ice.Message, pod, index string, args ...ice.Any) *ice.Message {
	if script := m.Cmdx(Space(m, pod), nfs.CAT, kit.Select(ice.SRC_MAIN_JS, index)); script != "" {
		return m.Echo(kit.Format(_main_template, script)).RenderResult()
	}
	return RenderIndex(m, ice.VOLCANOS)
}

var _cmd_template = `<!DOCTYPE html>
<head>
	<meta name="viewport" content="width=device-width,initial-scale=0.8,maximum-scale=0.8,user-scalable=no">
	<meta charset="utf-8">
	<link rel="stylesheet" type="text/css" href="/page/can.css">
</head>
<body>
	<script src="/page/can.js"></script><script>Volcanos(%s)</script>
</body>
`
var _main_template = `<!DOCTYPE html>
<head>
	<meta name="viewport" content="width=device-width,initial-scale=0.8,maximum-scale=0.8,user-scalable=no"/>
	<meta charset="utf-8">
	<title>volcanos</title>
	<link rel="shortcut icon" type="image/ico" href="/favicon.ico">
	<link rel="stylesheet" type="text/css" href="/page/cache.css">
	<link rel="stylesheet" type="text/css" href="/page/index.css">
</head>
<body>
	<script src="/proto.js"></script>
	<script src="/page/cache.js"></script>
	<script>%s</script>
</body>
`
