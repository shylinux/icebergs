package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"
	"github.com/skip2/go-qrcode"

	"fmt"
	"net/http"
	"path"
	"strings"
	"time"
)

const (
	REDIRECT = "redirect"
	REFRESH  = "refresh"
	STATUS   = "status"
	COOKIE   = "cookie"
)

func Render(msg *ice.Message, cmd string, args ...interface{}) {
	if cmd != "" {
		defer func() { msg.Log_EXPORT(cmd, args) }()
	}

	switch arg := kit.Simple(args...); cmd {
	case REDIRECT: // url [arg...]
		http.Redirect(msg.W, msg.R, kit.MergeURL(arg[0], arg[1:]), 307)

	case REFRESH: // [delay [text]]
		arg = []string{"200", fmt.Sprintf(`<!DOCTYPE html><head><meta charset="utf-8"><meta http-equiv="Refresh" content="%d"></head><body>%s</body>`,
			kit.Int(kit.Select("3", arg, 0)), kit.Select("请稍后，系统初始化中...", arg, 1),
		)}
		fallthrough

	case STATUS: // [code [text]]
		RenderStatus(msg, kit.Int(kit.Select("200", arg, 0)), kit.Select("", arg, 1))

	case COOKIE: // value [name [path [expire]]]
		RenderCookie(msg, arg[0], arg[1:]...)

	case ice.RENDER_DOWNLOAD: // file [type [name]]
		msg.W.Header().Set("Content-Disposition", fmt.Sprintf("filename=%s", kit.Select(path.Base(arg[0]), arg, 2)))
		if RenderType(msg.W, arg[0], kit.Select("", arg, 1)); !ice.DumpBinPack(msg.W, arg[0], nil) {
			http.ServeFile(msg.W, msg.R, kit.Path(arg[0]))
		}

	case ice.RENDER_QRCODE: // text [size]
		if qr, e := qrcode.New(arg[0], qrcode.Medium); msg.Assert(e) {
			msg.W.Header().Set(ContentType, ContentPNG)
			msg.Assert(qr.Write(kit.Int(kit.Select("256", arg, 1)), msg.W))
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

	default:
		for _, k := range []string{
			"_option", "_handle", "_output", "",
			"sessid", "domain", "river", "storm", "cmds", "fields",
		} {
			msg.Set(k)
		}

		if cmd != "" { // [str [arg...]]
			msg.Echo(kit.Format(cmd, args...))
		}
		msg.W.Header().Set(ContentType, ContentJSON)
		fmt.Fprint(msg.W, msg.Formats(kit.MDB_META))
	}
}

func RenderStatus(msg *ice.Message, code int, text string) {
	msg.W.WriteHeader(code)
	msg.W.Write([]byte(text))
}
func RenderCookie(msg *ice.Message, value string, arg ...string) { // name path expire
	expire := time.Now().Add(kit.Duration(kit.Select(msg.Conf(aaa.SESS, "meta.expire"), arg, 2)))
	http.SetCookie(msg.W, &http.Cookie{Value: value, Name: kit.Select(ice.MSG_SESSID, arg, 0), Path: kit.Select("/", arg, 1), Expires: expire})
}
func RenderType(w http.ResponseWriter, name, mime string) {
	if mime != "" {
		w.Header().Set(ContentType, mime)
	} else if strings.HasSuffix(name, ".css") {
		w.Header().Set(ContentType, "text/css; charset=utf-8")
	} else {
		w.Header().Set(ContentType, ContentHTML)
	}
}
