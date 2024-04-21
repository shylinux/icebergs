package wind

import (
	"runtime"

	"shylinux.com/x/ice"
	"shylinux.com/x/icebergs/base/cli"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

type wind struct {
	ice.Hash
	list string `name:"list hash auto"`
}

func (s wind) List(m *ice.Message, arg ...string) {
	m.Echo("hello world")
}

func init() { ice.ChatCtxCmd(wind{}) }

func Run(arg ...string) string {
	go func() { ice.Runs(func() {}, "serve", "start") }()
	runtime.LockOSThread()
	wnd := ui.NewWindowMain(ui.WindowMainOpts().Title("Contexts").ClientArea(win.SIZE{Cx: 340, Cy: 80}))
	btnShow := ui.NewButton(wnd, ui.ButtonOpts().Text("&Open").Position(win.POINT{X: 0, Y: 0}))
	btnShow.On().BnClicked(func() { ice.Pulse.Cmd(cli.SYSTEM, "explorer", "http://localhost:9020") })
	wnd.RunAsMain()
	return ""
}
