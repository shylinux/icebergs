package app

import (
	"github.com/webview/webview"
	"shylinux.com/x/ice"
)

var ww *webview.WebView

type app struct {
	title string `name:"title text" help:"标题"`
	list  string `name:"list auto title" help:"应用"`
}

func (app app) Title(m *ice.Message, arg ...string) {
	(*w).SetTitle("contexts")
}

func (app app) List(m *ice.Message, arg ...string) {
}
func init() { ice.Cmd("web.chat.app", app{}) }
func Run() {
	w := webview.New(true)
	defer w.Destroy()
	ww = &w

	w.SetTitle("contexts")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("http://localhost:9020")
	w.Run()
}
