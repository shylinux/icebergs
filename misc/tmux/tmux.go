package tmux

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const TMUX = "tmux"

var Index = &ice.Context{Name: TMUX, Help: "工作台", Configs: map[string]*ice.Config{
	TMUX: {Name: TMUX, Help: "工作台", Value: kit.Data(
		cli.SOURCE, "http://mirrors.tencent.com/macports/distfiles/tmux/tmux-3.2.tar.gz",
	)},
}, Commands: map[string]*ice.Command{
	TMUX: {Name: "tmux path auto start order build download", Help: "服务", Action: ice.MergeAction(map[string]*ice.Action{
		cli.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
			m.Optionv(code.PREPARE, func(p string) []string {
				return []string{"-S", kit.Path(m.Option(cli.CMD_DIR, p), "tmux.socket"), "new-session", "-dn", "miss"}
			})
			m.Cmdy(code.INSTALL, cli.START, m.Config(cli.SOURCE), "bin/tmux")
		}},
	}, code.InstallAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Cmdy(code.INSTALL, cli.SOURCE, m.Config(cli.SOURCE), arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}) }
