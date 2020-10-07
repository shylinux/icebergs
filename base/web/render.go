package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	kit "github.com/shylinux/toolkits"
	"github.com/skip2/go-qrcode"

	"fmt"
	"net/http"
	"os"
	"path"
	"time"
)

const (
	STATUS = "status"
	COOKIE = "cookie"
)

func Render(msg *ice.Message, cmd string, args ...interface{}) {
	if cmd != "" {
		defer func() { msg.Log(ice.LOG_EXPORT, "%s: %v", cmd, args) }()
	}
	switch arg := kit.Simple(args...); cmd {
	case ice.RENDER_VOID:
	case ice.RENDER_OUTPUT:
	case "redirect":
		http.Redirect(msg.W, msg.R, kit.MergeURL(arg[0], arg[1:]), 307)

	case "refresh":
		arg = []string{"200", fmt.Sprintf(`<!DOCTYPE html><head><meta charset="utf-8"><meta http-equiv="Refresh" content="%d"></head><body>%s</body>`,
			kit.Int(kit.Select("3", arg, 0)), kit.Select("请稍后，系统初始化中...", arg, 1),
		)}
		fallthrough

	case STATUS:
		RenderStatus(msg, kit.Int(kit.Select("200", arg, 0)), kit.Select("", arg, 1))

	case COOKIE:
		RenderCookie(msg, arg[0], arg[1:]...)

	case ice.RENDER_DOWNLOAD:
		msg.W.Header().Set("Content-Disposition", fmt.Sprintf("filename=%s", kit.Select(path.Base(arg[0]), arg, 2)))
		msg.W.Header().Set("Content-Type", kit.Select("text/html", arg, 1))
		if _, e := os.Stat(arg[0]); e != nil {
			arg[0] = "/" + arg[0]
		}
		http.ServeFile(msg.W, msg.R, arg[0])

	case ice.RENDER_RESULT:
		if len(arg) > 0 {
			msg.W.Write([]byte(kit.Format(arg[0], args[1:]...)))
		} else {
			args = append(args, "length:", len(msg.Result()))
			msg.W.Write([]byte(msg.Result()))
		}

	case ice.RENDER_QRCODE:
		if qr, e := qrcode.New(arg[0], qrcode.Medium); msg.Assert(e) {
			msg.W.Header().Set("Content-Type", "image/png")
			msg.Assert(qr.Write(kit.Int(kit.Select("256", arg, 1)), msg.W))
		}

	default:
		if cmd != "" {
			msg.Echo(kit.Format(cmd, args...))
		}
		msg.W.Header().Set("Content-Type", "application/json")
		fmt.Fprint(msg.W, msg.Formats("meta"))
	}
	msg.Append(ice.MSG_OUTPUT, ice.RENDER_OUTPUT)
}

func RenderCookie(msg *ice.Message, value string, arg ...string) { // name path expire
	expire := time.Now().Add(kit.Duration(kit.Select(msg.Conf(aaa.SESS, "meta.expire"), arg, 2)))
	http.SetCookie(msg.W, &http.Cookie{Value: value, Name: kit.Select(ice.MSG_SESSID, arg, 0), Path: kit.Select("/", arg, 1), Expires: expire})
}
func RenderStatus(msg *ice.Message, code int, text string) {
	msg.W.WriteHeader(code)
	msg.W.Write([]byte(text))
}
