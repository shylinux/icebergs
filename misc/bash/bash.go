package bash

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const BASH = "bash"

var Index = &ice.Context{Name: BASH, Help: "命令行",
	Configs: map[string]*ice.Config{
		BASH: {Name: BASH, Help: "命令行", Value: kit.Data(
			cli.SOURCE, "http://mirrors.aliyun.com/gnu/bash/bash-4.4.18.tar.gz",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		BASH: {Name: "bash port path auto start build download", Help: "命令行", Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, web.DOWNLOAD, m.Conf(BASH, kit.Keym(cli.SOURCE)))
			}},
			cli.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, cli.BUILD, m.Conf(BASH, kit.Keym(cli.SOURCE)))
			}},
			cli.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, cli.START, m.Conf(BASH, kit.Keym(cli.SOURCE)), "bin/bash")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(BASH, kit.Keym(cli.SOURCE))), arg)
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
