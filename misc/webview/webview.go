package webview

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/webview"
)

const (
	CONF_SIZE = "var/conf/webview.size"
)

type WebView struct {
	webview.WebView
	Source string
	*ice.Message
}

func (w WebView) Menu() bool {
	link, list := "", []string{}
	w.Cmd(nfs.CAT, w.Source, func(ls []string, line string) {
		if strings.HasPrefix(line, "# ") {
			return
		} else if len(ls) > 1 {
			link, list = ls[1], append(list, kit.Format(`<button onclick=%s()>%s</button>`, ls[0], ls[0]))
			w.WebView.Bind(ls[0], func() { w.navigate(ls[1]) })
		}
	})
	if len(list) == 0 {
		return false
	} else if len(list) == 1 {
		if ls := kit.Split(w.Cmdx(nfs.CAT, CONF_SIZE)); len(ls) > 1 {
			w.WebView.SetSize(kit.Int(ls[0]), kit.Int(ls[1])+28, webview.HintNone)
		} else {
			w.WebView.SetSize(1200, 800, webview.HintNone)
		}
		w.WebView.Navigate(link)
		return true
	} else {
		w.WebView.SetTitle(ice.CONTEXTS)
		w.WebView.SetSize(200, 60*len(list), webview.HintNone)
		w.WebView.Navigate(kit.Format(`data:text/html,`+ice.Pulse.Cmdx(nfs.CAT, path.Join(ice.SRC_TEMPLATE, "webview", "home.html")), kit.Join(list, lex.NL)))
		return true
	}
}
func (w WebView) Title(text string)  { w.WebView.SetTitle(text) }
func (w WebView) Webview(url string) { w.WebView.Navigate(url) }
func (w WebView) Open(url string) {
	w.Message.Debug("open %v", url)
	w.WebView.Navigate(url)
}
func (w WebView) OpenUrl(url string) {
	w.Message.Debug("open %v", url)
	cli.Opens(w.Message, url)
}
func (w WebView) OpenApp(app string) {
	w.Message.Debug("open %v", app)
	cli.Opens(w.Message, app)
}
func (w WebView) OpenCmd(cmd string) {
	w.Cmd(nfs.SAVE, kit.HomePath(".bash_temp"), cmd)
	cli.Opens(w.Message, "Terminal.app", "-n")
}
func (w WebView) SetSize(width, height int) {
	w.Cmd(nfs.SAVE, CONF_SIZE, kit.Format("%v,%v", width, height))
}
func (w WebView) System(arg ...string) string { return w.Cmdx(cli.SYSTEM, arg) }
func (w WebView) Power() string {
	ls := strings.Split(w.Cmdx(cli.SYSTEM, "pmset", "-g", "ps"), lex.NL)
	for _, line := range ls[1:] {
		ls := kit.Split(line, "\t ;", "\t ;")
		return ls[2]
	}
	return ""
}
func (w WebView) Close() { kit.If(!w.Menu(), func() { w.WebView.Terminate() }) }
func (w WebView) Terminate() {
	w.WebView.Eval("window.onbeforeunload()")
	w.WebView.Terminate()
}
func (w WebView) navigate(url string) {
	w.WebView.SetSize(1200, 800, webview.HintNone)
	w.WebView.Navigate(url)
}

func Run(cb func(*WebView) ice.Any) {
	w := webview.New(true)
	defer w.Destroy()
	defer w.Run()
	view := &WebView{Source: "etc/webview.txt", WebView: w, Message: ice.Pulse.Spawn(kit.Dict(ice.MSG_USERIP, tcp.LOCALHOST))}
	if cb == nil {
		kit.Reflect(view, func(name string, value ice.Any) { w.Bind(name, value) })
	} else {
		kit.Reflect(cb(view), func(name string, value ice.Any) { w.Bind(name, value) })
	}
	kit.If(!view.Menu(), func() { view.navigate("http://localhost:9020") })
}
