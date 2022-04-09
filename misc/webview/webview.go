package webview

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/webview"
)

type WebView struct {
	Source  string
	Target  interface{}
	WebView webview.WebView
}

func (w WebView) Menu() bool {
	kit.Reflect(w.Target, func(name string, value interface{}) { w.WebView.Bind(name, value) })
	list := []string{}
	ice.Pulse.Cmd(nfs.CAT, w.Source, func(ls []string, line string) {
		if len(ls) > 1 {
			list = append(list, kit.Format(`<button onclick=%s()>%s</button>`, ls[0], ls[0]))
			w.WebView.Bind(ls[0], func() {
				w.WebView.SetSize(1200, 800, webview.HintNone)
				w.WebView.Navigate(ls[1])
			})
		}
	})

	if len(list) == 0 {
		return false
	}

	w.WebView.SetTitle("contexts")
	w.WebView.SetSize(200, 60*len(list), webview.HintNone)
	w.WebView.Navigate(kit.Format(`data:text/html,
    <!doctype html>
    <html>
	<head>
	<style>button { font-size:24px; font-family:monospace; margin:10px; width:-webkit-fill-available; display:block; clear:both; }</style>
	<script>
window.alert("hello world")
document.body.onclick = function(event) {
	if (event.metaKey) {
		switch (event.key) {
			case "w": close() break
			case "q": terminate() break
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
func (w WebView) Terminate()         { w.WebView.Terminate() }
func (w WebView) Close() {
	if !w.Menu() {
		w.WebView.Terminate()
	}
}

func Run(cb func(*WebView) interface{}) {
	w := webview.New(true)
	defer w.Destroy()
	defer w.Run()

	w.Init(`
window.alert("hello world")
document.body.onclick = function(event) {
	if (event.metaKey) {
		switch (event.key) {
			case "w": close() break
			case "q": terminate() break
		}
	}
}
`)

	view := &WebView{Source: "src/webview.txt", WebView: w}
	target := cb(view)

	if view.Target = target; !view.Menu() {
		w.SetSize(1200, 800, webview.HintNone)
		w.Navigate("http://localhost:9020")
	}
}
