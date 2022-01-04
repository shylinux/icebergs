package app

import (
	"github.com/webview/webview"
	"shylinux.com/x/ice"
	kit "shylinux.com/x/toolkits"
)

type app struct {
	title string `name:"title text" help:"标题"`
	list  string `name:"list auto title" help:"应用"`
}

func (app app) Title(m *ice.Message, arg ...string) {
	(*ww).SetTitle("contexts")
}

func (app app) List(m *ice.Message, arg ...string) {
}
func init() { ice.Cmd("web.chat.app", app{}) }

var ww *webview.WebView

func Run(arg ...string) {
	w := webview.New(true)
	defer w.Destroy()
	ww = &w

	w.SetSize(800, 600, webview.HintNone)
	w.SetTitle(kit.Select("contexts", arg, 0))
	w.Navigate(kit.Select("http://localhost:9020", arg, 1))
	w.Run()
}
