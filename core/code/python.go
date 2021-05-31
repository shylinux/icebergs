package code

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	kit "github.com/shylinux/toolkits"
)

const PYTHON = "python"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			PYTHON: {Name: "python", Help: "脚本命令", Value: kit.Data(
				PYTHON, "python", cli.SOURCE, "http://mirrors.sohu.com/python/3.5.2/Python-3.5.2.tar.xz",
			)},
		},
		Commands: map[string]*ice.Command{
			PYTHON: {Name: "python port path auto start build download", Help: "脚本命令", Action: map[string]*ice.Action{
				"download": {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.install", "download", m.Conf(PYTHON, "meta.source"))
				}},
				"build": {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.install", "build", m.Conf(PYTHON, "meta.source"))
				}},
				"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy("web.code.install", "start", m.Conf(PYTHON, "meta.source"), "bin/python3")
				}},

				"pip": {Name: "pip", Help: "安装", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, m.Conf(PYTHON, "meta.pip"), "install", arg)
				}},
				"run": {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, m.Conf(PYTHON, "meta.python"), arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy("web.code.install", path.Base(m.Conf(PYTHON, kit.META_SOURCE)), arg)
			}},
		},
	})
}
