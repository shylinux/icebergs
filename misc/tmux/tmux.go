package tmux

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

func _tmux_cmd(m *ice.Message, arg ...string) *ice.Message { return m.Cmd(cli.SYSTEM, TMUX, arg) }
func _tmux_cmds(m *ice.Message, arg ...string) string      { return _tmux_cmd(m, arg...).Results() }

const TMUX = "tmux"

var Index = &ice.Context{Name: TMUX, Help: "工作台", Commands: ice.Commands{
	TMUX: {Help: "服务", Actions: ice.MergeActions(ice.Actions{
		cli.START: {Hand: func(m *ice.Message, arg ...string) {
			m.Optionv(code.PREPARE, func(p string) []string {
				return []string{"-S", kit.Path(m.Option(cli.CMD_DIR, p), "tmux.socket"), NEW_SESSION, "-d", "-n", "miss"}
			})
			m.Cmdy(code.INSTALL, cli.START, mdb.Config(m, nfs.SOURCE), "bin/tmux")
		}},
	}, code.InstallAction(nfs.SOURCE, "http://mirrors.tencent.com/macports/distfiles/tmux/tmux-3.2.tar.gz"))},
}}

func init() { code.Index.Register(Index, &web.Frame{}) }
