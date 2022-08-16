package tmux

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const TMUX = "tmux"

var Index = &ice.Context{Name: TMUX, Help: "工作台", Commands: ice.Commands{
	TMUX: {Name: "tmux path auto start order build download", Help: "服务", Actions: ice.MergeActions(ice.Actions{
		cli.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
			m.Optionv(code.PREPARE, func(p string) []string {
				return []string{"-S", kit.Path(m.Option(cli.CMD_DIR, p), "tmux.socket"), NEW_SESSION, "-d", "-n", "miss"}
			})
			m.Cmdy(code.INSTALL, cli.START, m.Config(nfs.SOURCE), "bin/tmux")
		}},
	}, code.InstallAction(nfs.SOURCE, "http://mirrors.tencent.com/macports/distfiles/tmux/tmux-3.2.tar.gz")), Hand: func(m *ice.Message, arg ...string) {
		m.Cmdy(code.INSTALL, nfs.SOURCE, m.Config(nfs.SOURCE), arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}) }
