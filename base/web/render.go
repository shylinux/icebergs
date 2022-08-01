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
		defer func() { msg.Log_EXPORT(cmd, args) }()
	}

	switch arg := kit.Simple(args...); cmd {
	case COOKIE: // value [name [path [expire]]]
		RenderCookie(msg, arg[0], arg[1:]...)

	case STATUS, ice.RENDER_STATUS: // [code [text]]
		RenderStatus(msg, kit.Int(kit.Select("200", arg, 0)), kit.Select("", arg, 1))

	case ice.RENDER_REDIRECT: // url [arg...]
		RenderRedirect(msg, arg...)

	case ice.RENDER_DOWNLOAD: // file [type [name]]
		if strings.HasPrefix(arg[0], ice.HTTP) {
			http.Redirect(msg.W, msg.R, arg[0], http.StatusSeeOther)
			break
		}
		msg.W.Header().Set("Content-Disposition", fmt.Sprintf("filename=%s", kit.Select(path.Base(kit.Select(arg[0], msg.Option("filename"))), arg, 2)))
		RenderType(msg.W, arg[0], kit.Select("", arg, 1))
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
		msg.W.Header().Set("Content-Type", "application/json")
		msg.W.Write([]byte(arg[0]))

	case ice.RENDER_VOID:
		// no output

	case ice.RENDER_RAW:
		fallthrough
	default:
		for _, k := range []string{
			"_", "_option", "_handle", "_output", "",
			"cmds", "fields", "sessid", "domain",
			"river", "storm",
		} {
			msg.Set(k)
		}

		if cmd != "" && cmd != ice.RENDER_RAW { // [str [arg...]]
			msg.Echo(kit.Format(cmd, args...))
		}
		msg.W.Header().Set(ContentType, ContentJSON)
		fmt.Fprint(msg.W, msg.FormatMeta())
	}
}
func RenderType(w http.ResponseWriter, name, mime string) {
	if mime != "" {
		w.Header().Set(ContentType, mime)
		return
	}

	switch kit.Ext(name) {
	case nfs.CSS:
		w.Header().Set(ContentType, "text/css; charset=utf-8")
	case "pdf":
		w.Header().Set(ContentType, "application/pdf")
	}
}
func RenderHeader(msg *ice.Message, key, value string) {
	msg.W.Header().Set(key, value)
}
func RenderCookie(msg *ice.Message, value string, arg ...string) { // name path expire
	expire := time.Now().Add(kit.Duration(kit.Select(msg.Conf(aaa.SESS, kit.Keym(mdb.EXPIRE)), arg, 2)))
	http.SetCookie(msg.W, &http.Cookie{Value: value,
		Name: kit.Select(CookieName(msg.Option(ice.MSG_USERWEB)), arg, 0), Path: kit.Select(ice.PS, arg, 1), Expires: expire})
}
func RenderStatus(msg *ice.Message, code int, text string) {
	msg.W.WriteHeader(code)
	msg.W.Write([]byte(text))
}
func RenderRefresh(msg *ice.Message, arg ...string) { // url text delay
	msg.Render(ice.RENDER_VOID)
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
}
func RenderRedirect(msg *ice.Message, arg ...string) {
	http.Redirect(msg.W, msg.R, kit.MergeURL(arg[0], arg[1:]), http.StatusTemporaryRedirect)
}
func RenderDownload(msg *ice.Message, arg ...ice.Any) {
	Render(msg, ice.RENDER_DOWNLOAD, arg...)
}
func RenderResult(msg *ice.Message, arg ...ice.Any) {
	Render(msg, ice.RENDER_RESULT, arg...)
}
func CookieName(url string) string {
	return ice.MSG_SESSID + "_" + kit.ReplaceAll(kit.ParseURLMap(url)[tcp.HOST], ".", "_", ":", "_")
}
func Format(tag string, arg ...ice.Any) string {
	return kit.Format("<%s>%s</%s>", tag, strings.Join(kit.Simple(arg), ""), tag)
}
