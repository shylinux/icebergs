package webview

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/webview"
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
		if ls := kit.Split(w.Cmdx(nfs.CAT, "etc/webview.size")); len(ls) > 1 {
			w.WebView.SetSize(kit.Int(ls[0]), kit.Int(ls[1])+28, webview.HintNone)
		} else {
			w.WebView.SetSize(1200, 800, webview.HintNone)
		}
		w.WebView.Navigate(link)
		return true
	} else {
		w.WebView.SetTitle(ice.CONTEXTS)
		w.WebView.SetSize(200, 60*len(list), webview.HintNone)
		w.WebView.Navigate(kit.Format(_menu_template, kit.Join(list, ice.NL)))
		return true
	}
}
func (w WebView) Title(text string)  { w.WebView.SetTitle(text) }
func (w WebView) Webview(url string) { w.WebView.Navigate(url) }
func (w WebView) Open(url string)    { w.WebView.Navigate(url) }
func (w WebView) OpenUrl(url string) { w.Cmd(cli.SYSTEM, cli.OPEN, url) }
func (w WebView) OpenApp(app string) { w.Cmd(cli.SYSTEM, cli.OPEN, "-a", app) }
func (w WebView) OpenCmd(cmd string) {
	w.Cmd(nfs.SAVE, kit.HomePath(".bash_temp"), cmd)
	w.Cmd(cli.SYSTEM, cli.OPEN, "-a", "Terminal")
}
func (w WebView) SetSize(width, height int) {
	w.Cmd(nfs.SAVE, "etc/webview.size", kit.Format("%v,%v", width, height))
}
func (w WebView) System(arg ...string) string { return w.Cmdx(cli.SYSTEM, arg) }
func (w WebView) Power() string {
	ls := strings.Split(w.Cmdx(cli.SYSTEM, "pmset", "-g", "ps"), ice.NL)
	for _, line := range ls[1:] {
		ls := kit.Split(line, "\t ;", "\t ;")
		return ls[2]
	}
	return ""
}
func (w WebView) Close() {
	if !w.Menu() {
		w.WebView.Terminate()
	}
}
func (w WebView) Terminate() { w.WebView.Terminate() }
func (w WebView) navigate(url string) {
	w.WebView.SetSize(1200, 800, webview.HintNone)
	w.WebView.Navigate(url)
}

func Run(cb func(*WebView) ice.Any) {
	w := webview.New(true)
	defer w.Destroy()
	defer w.Run()
	view := &WebView{Source: "etc/webview.txt", WebView: w, Message: ice.Pulse}
	kit.Reflect(cb(view), func(name string, value ice.Any) { w.Bind(name, value) })
	if !view.Menu() {
		view.navigate("http://localhost:9020")
	}
}

var _menu_template = `data:text/html,
<!doctype html>
<html>
<head>
	<style>button { font-size:24px; font-family:monospace; margin:10px; width:-webkit-fill-available; display:block; clear:both; }</style>
	<script>
		document.body.onkeydown = function(event) {
			if (event.metaKey) {
				switch (event.key) {
				case "q": window.terminate(); break
				}
			}
		}
	</script>
</head>
<body>%s</body>
</html>`
