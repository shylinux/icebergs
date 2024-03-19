package webview

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web"
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
		if len(ls) == 0 || strings.HasPrefix(line, "# ") {
			return
		} else if len(ls) == 1 {
			u := kit.ParseURL(ls[0])
			ls = []string{strings.ReplaceAll(u.Hostname(), ".", "_"), ls[0]}
		}
		w.WebView.Bind(ls[0], func() { w.navigate(ls[1]) })
		link, list = ls[1], append(list, kit.Format(`<button onclick=%s()>%s</button>`, ls[0], ls[0]))
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
func (w WebView) Open(url string)    { w.WebView.Navigate(url) }
func (w WebView) OpenUrl(url string) { cli.Opens(w.Message, url) }
func (w WebView) OpenApp(app string) { cli.Opens(w.Message, app) }
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
func (w WebView) Close()     { kit.If(!w.Menu(), func() { w.WebView.Terminate() }) }
func (w WebView) Terminate() { w.WebView.Terminate() }
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
	kit.If(!view.Menu(), func() { view.navigate(ice.Pulse.Cmdv(web.SPIDE, ice.OPS, web.CLIENT_ORIGIN)) })
}
func RunClient() {
	kit.Setenv(cli.PATH, "/usr/local/bin:/usr/bin:/bin:/usr/local/sbin:/usr/sbin:/sbin")
	kit.Chdir(kit.HomePath(ice.CONTEXTS))
	wait := make(chan bool, 1)
	ice.Pulse.Optionv(web.SERVE_START, func() { wait <- true })
	go ice.Run(ice.SERVE, ice.START)
	defer ice.Pulse.Cmd(ice.EXIT)
	defer Run(nil)
	<-wait
}
