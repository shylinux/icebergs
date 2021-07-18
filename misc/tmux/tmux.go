package tmux

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

const (
	TEXT    = "text"
	BUFFER  = "buffer"
	SCRIPT  = "script"
	SESSION = "session"
	WINDOW  = "window"
	PANE    = "pane"
	VIEW    = "view"
)

const TMUX = "tmux"

var Index = &ice.Context{Name: TMUX, Help: "工作台",
	Configs: map[string]*ice.Config{
		TMUX: {Name: TMUX, Help: "服务", Value: kit.Data(
			cli.SOURCE, "https://github.com/tmux/tmux/releases/download/3.1b/tmux-3.1b.tar.gz",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		TMUX: {Name: "tmux port path auto start build download", Help: "服务", Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, web.DOWNLOAD, m.Conf(TMUX, kit.Keym(cli.SOURCE)))
			}},
			cli.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, cli.BUILD, m.Conf(TMUX, kit.Keym(cli.SOURCE)))
			}},
			cli.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Optionv(code.PREPARE, func(p string) []string {
					return []string{"-S", kit.Path(m.Option(cli.CMD_DIR, p), "tmux.socket"), "new-session", "-dn", "miss"}
				})
				m.Cmdy(code.INSTALL, cli.START, m.Conf(TMUX, kit.Keym(cli.SOURCE)), "bin/tmux")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(TMUX, kit.Keym(cli.SOURCE))), arg)
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
