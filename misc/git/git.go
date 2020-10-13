package git

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
)

const GIT = "git"

var Index = &ice.Context{Name: GIT, Help: "代码库",
	Configs: map[string]*ice.Config{
		GIT: {Name: GIT, Help: "代码库", Value: kit.Data(
			"source", "https://mirrors.edge.kernel.org/pub/software/scm/git/git-1.8.3.1.tar.gz", "config", kit.Dict(
				"alias", kit.Dict("s", "status", "b", "branch"),
				"color", kit.Dict("ui", "true"),
				"push", kit.Dict("default", "simple"),
				"credential", kit.Dict("helper", "store"),
			),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 系统项目
			wd, _ := os.Getwd()
			_repos_insert(m, path.Base(wd), wd)

			m.Cmd(nfs.DIR, "usr", "name path").Table(func(index int, value map[string]string, head []string) {
				_repos_insert(m, value["name"], value["path"])
			})
		}},
		"init": {Name: "init", Help: "初始化", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 官方项目
			// m.Cmd(nfs.DIR, "usr", "name path").Table(func(index int, value map[string]string, head []string) {
			// 	_repos_insert(m, value["name"], value["path"])
			// })
		}},

		GIT: {Name: "git port=auto path=auto auto 启动 构建 下载", Help: "代码库", Action: map[string]*ice.Action{
			"download": {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, "download", m.Conf(GIT, kit.META_SOURCE))
			}},
			"build": {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, "build", m.Conf(GIT, kit.META_SOURCE))
			}},
			"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Optionv("prepare", func(p string) []string {
					m.Option(cli.CMD_DIR, p)
					// kit.Fetch(m.Confv(GIT, "meta.config"), func(conf string, value interface{}) {
					// 	kit.Fetch(value, func(key string, value string) {
					// 		m.Cmd(cli.SYSTEM, "bin/git", "config", "--global", conf+"."+key, value)
					// 	})
					// })
					return []string{}
				})
				m.Cmdy(code.INSTALL, "start", m.Conf(GIT, kit.META_SOURCE), "bin/git")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(GIT, kit.META_SOURCE)), arg)
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
