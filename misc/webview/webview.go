package webview

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/webview"
)

type WebView struct {
	webview.WebView
	Source string
}

func (w WebView) Menu() bool {
	list := []string{}
	ice.Pulse.Cmd(nfs.CAT, w.Source, func(ls []string, line string) {
		if len(ls) > 1 {
			list = append(list, kit.Format(`<button onclick=%s()>%s</button>`, ls[0], ls[0]))
			w.WebView.Bind(ls[0], func() { w.navigate(ls[1]) })
		}
	})
	if len(list) == 0 {
		return false
	}

	w.WebView.SetTitle(ice.CONTEXTS)
	w.WebView.SetSize(200, 60*len(list), webview.HintNone)
	w.WebView.Navigate(kit.Format(`data:text/html,
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
</html>`, kit.Join(list, ice.NL)))
	return true
}
func (w WebView) Title(text string)  { w.WebView.SetTitle(text) }
func (w WebView) Webview(url string) { w.WebView.Navigate(url) }
func (w WebView) Open(url string)    { w.WebView.Navigate(url) }
func (w WebView) OutOpen(url string) { ice.Pulse.Cmd("system", "open", url) }
func (w WebView) Terminal(url string) { ice.Pulse.Cmd("system", "open", "-a", "Terminal") }
func (w WebView) Terminate()         { w.WebView.Terminate() }
func (w WebView) Close() {
	if !w.Menu() {
		w.WebView.Terminate()
	}
}
func (w WebView) navigate(url string) {
	w.WebView.SetSize(1200, 800, webview.HintNone)
	w.WebView.Navigate(url)
}

func Run(cb func(*WebView) ice.Any) {
	w := webview.New(true)
	defer w.Destroy()
	defer w.Run()

	view := &WebView{Source: "src/webview.txt", WebView: w}
	kit.Reflect(cb(view), func(name string, value ice.Any) { w.Bind(name, value) })

	if !view.Menu() {
		view.navigate("http://localhost:9020")
	}
}
