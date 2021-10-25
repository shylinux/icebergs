package bash

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const BASH = "bash"

var Index = &ice.Context{Name: BASH, Help: "命令行", Configs: map[string]*ice.Config{
	BASH: {Name: BASH, Help: "命令行", Value: kit.Data(
		cli.SOURCE, "http://mirrors.tencent.com/macports/distfiles/bash/5.1_1/bash-5.1.tar.gz",
	)},
}, Commands: map[string]*ice.Command{
	BASH: {Name: "bash path auto order build download", Help: "命令行", Action: ice.MergeAction(map[string]*ice.Action{
		cli.ORDER: {Name: "order", Help: "加载", Hand: func(m *ice.Message, arg ...string) {
			m.Cmd(code.INSTALL, cli.ORDER, m.Config(cli.SOURCE), "_install/bin")
		}},
	}, code.InstallAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Cmdy(code.INSTALL, cli.SOURCE, m.Config(cli.SOURCE), arg)
	}},
}}

func init() { code.Index.Register(Index, &web.Frame{}) }
