package code

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const PYTHON = "python"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PYTHON: {Name: PYTHON, Help: "脚本命令", Value: kit.Data(
				cli.SOURCE, "http://mirrors.sohu.com/python/3.5.2/Python-3.5.2.tar.xz",
				PYTHON, "python",
			)},
		},
		Commands: map[string]*ice.Command{
			PYTHON: {Name: "python port path auto start build download", Help: "脚本命令", Action: map[string]*ice.Action{
				web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(INSTALL, web.DOWNLOAD, m.Conf(PYTHON, kit.Keym(cli.SOURCE)))
				}},
				cli.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(INSTALL, cli.BUILD, m.Conf(PYTHON, kit.Keym(cli.SOURCE)))
				}},
				cli.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(INSTALL, cli.START, m.Conf(PYTHON, kit.Keym(cli.SOURCE)), "bin/python3")
				}},
				cli.RUN: {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, m.Conf(PYTHON, kit.Keym(PYTHON)), arg)
				}},
				"pip": {Name: "pip", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, m.Conf(PYTHON, kit.Keym("pip")), "install", arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(INSTALL, path.Base(m.Conf(PYTHON, kit.META_SOURCE)), arg)
			}},
		},
	})
}
