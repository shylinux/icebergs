package web

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

const (
	STATUS = "status"
	COOKIE = "cookie"
)

func Render(msg *ice.Message, cmd string, args ...interface{}) {
	if cmd != "" {
		defer func() { msg.Log_EXPORT(cmd, args) }()
	}

	switch arg := kit.Simple(args...); cmd {
	case STATUS: // [code [text]]
		RenderStatus(msg, kit.Int(kit.Select("200", arg, 0)), kit.Select("", arg, 1))

	case COOKIE: // value [name [path [expire]]]
		RenderCookie(msg, arg[0], arg[1:]...)

	case ice.RENDER_REDIRECT: // url [arg...]
		RenderRedirect(msg, arg...)

	case ice.RENDER_DOWNLOAD: // file [type [name]]
		if strings.HasPrefix(arg[0], "http") {
			http.Redirect(msg.W, msg.R, arg[0], http.StatusSeeOther)
			break
		}
		msg.W.Header().Set("Content-Disposition", fmt.Sprintf("filename=%s", kit.Select(path.Base(kit.Select(arg[0], msg.Option("filename"))), arg, 2)))
		if RenderType(msg.W, arg[0], kit.Select("", arg, 1)); !ice.Dump(msg.W, arg[0], nil) {
			http.ServeFile(msg.W, msg.R, kit.Path(arg[0]))
		}

	case ice.RENDER_RESULT:
		if len(arg) > 0 { // [str [arg...]]
			msg.W.Write([]byte(kit.Format(arg[0], args[1:]...)))
		} else {
			args = append(args, "length:", len(msg.Result()))
			msg.W.Write([]byte(msg.Result()))
		}

	case ice.RENDER_VOID:
		// no output

	case ice.RENDER_RAW:
		fallthrough
	default:
		for _, k := range []string{
			"_option", "_handle", "_output", "",
			"cmds", "fields", "sessid", "domain",
			"river", "storm",
		} {
			msg.Set(k)
		}

		if cmd != "" && cmd != ice.RENDER_RAW { // [str [arg...]]
			msg.Echo(kit.Format(cmd, args...))
		}
		msg.W.Header().Set(ContentType, ContentJSON)
		fmt.Fprint(msg.W, msg.FormatsMeta())
	}
}
func RenderHeader(msg *ice.Message, key, value string) {
	msg.W.Header().Set(key, value)
}
func RenderStatus(msg *ice.Message, code int, text string) {
	msg.W.WriteHeader(code)
	msg.W.Write([]byte(text))
}
func CookieName(url string) string {
	return ice.MSG_SESSID + "_" + kit.ReplaceAll(kit.ParseURLMap(url)[tcp.HOST], ".", "_", ":", "_")
}
func RenderCookie(msg *ice.Message, value string, arg ...string) { // name path expire
	expire := time.Now().Add(kit.Duration(kit.Select(msg.Conf(aaa.SESS, "meta.expire"), arg, 2)))
	http.SetCookie(msg.W, &http.Cookie{Value: value,
		Name: kit.Select(CookieName(msg.Option(ice.MSG_USERWEB)), arg, 0), Path: kit.Select(ice.PS, arg, 1), Expires: expire})
}
func RenderRedirect(msg *ice.Message, arg ...string) {
	http.Redirect(msg.W, msg.R, kit.MergeURL(arg[0], arg[1:]), http.StatusMovedPermanently)
}
func RenderType(w http.ResponseWriter, name, mime string) {
	if mime != "" {
		w.Header().Set(ContentType, mime)
		return
	}

	switch kit.Ext(name) {
	case "css":
		w.Header().Set(ContentType, "text/css; charset=utf-8")
	case "pdf":
		w.Header().Set(ContentType, "application/pdf")
	default:
	}
}
func RenderResult(msg *ice.Message, arg ...interface{}) {
	Render(msg, ice.RENDER_RESULT, arg...)
}
func RenderDownload(msg *ice.Message, arg ...interface{}) {
	Render(msg, ice.RENDER_DOWNLOAD, arg...)
}

type Buffer struct {
	m *ice.Message
	n string
}

func (b *Buffer) Write(buf []byte) (int, error) {
	b.m.PushNoticeGrow(string(buf))
	return len(buf), nil
}
func (b *Buffer) Close() error { return nil }

func PushStream(m *ice.Message) {
	m.Option(cli.CMD_OUTPUT, &Buffer{m: m, n: m.Option(ice.MSG_DAEMON)})
}

func Format(tag string, arg ...interface{}) string {
	return kit.Format("<%s>%s</%s>", tag, strings.Join(kit.Simple(arg), ""), tag)
}
