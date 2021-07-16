package git

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

const GIT = "git"

var Index = &ice.Context{Name: GIT, Help: "代码库",
	Configs: map[string]*ice.Config{
		GIT: {Name: GIT, Help: "代码库", Value: kit.Data(
			cli.SOURCE, "https://mirrors.edge.kernel.org/pub/software/scm/git/git-1.8.3.1.tar.gz",
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

		GIT: {Name: "git port path auto start build download", Help: "源代码", Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, web.DOWNLOAD, m.Conf(GIT, kit.Keym(cli.SOURCE)))
			}},
			cli.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, cli.BUILD, m.Conf(GIT, kit.Keym(cli.SOURCE)))
			}},
			cli.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, cli.START, m.Conf(GIT, kit.Keym(cli.SOURCE)), "bin/git")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(GIT, kit.Keym(cli.SOURCE))), arg)
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
