package git

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

const GIT = "git"

var Index = &ice.Context{Name: GIT, Help: "代码库",
	Configs: map[string]*ice.Config{
		GIT: {Name: GIT, Help: "代码库", Value: kit.Data(
			kit.SSH_SOURCE, "https://mirrors.edge.kernel.org/pub/software/scm/git/git-1.8.3.1.tar.gz", "config", kit.Dict(
				"alias", kit.Dict("s", "status", "b", "branch"),
				"credential", kit.Dict("helper", "store"),
				"core", kit.Dict("quotepath", "false"),
				"push", kit.Dict("default", "simple"),
				"color", kit.Dict("ui", "always"),
			),
		)},
	},
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_repos_insert(m, path.Base(kit.Pwd()), kit.Pwd())

			m.Cmd(nfs.DIR, kit.SSH_USR, "name,path").Table(func(index int, value map[string]string, head []string) {
				_repos_insert(m, value[kit.MDB_NAME], value[kit.MDB_PATH])
			})
		}},

		GIT: {Name: "git port path auto start build download", Help: "代码库", Action: map[string]*ice.Action{
			web.DOWNLOAD: {Name: "download", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, web.DOWNLOAD, m.Conf(GIT, kit.META_SOURCE))
			}},
			gdb.BUILD: {Name: "build", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.INSTALL, gdb.BUILD, m.Conf(GIT, kit.META_SOURCE))
			}},
			gdb.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				m.Optionv(code.PREPARE, func(p string) []string {
					m.Option(cli.CMD_DIR, p)
					kit.Fetch(m.Confv(GIT, kit.Keym("config")), func(conf string, value interface{}) {
						kit.Fetch(value, func(key string, value string) {
							m.Cmd(cli.SYSTEM, "bin/git", "config", "--global", conf+"."+key, value)
						})
					})
					return []string{}
				})
				m.Cmdy(code.INSTALL, gdb.START, m.Conf(GIT, kit.META_SOURCE), "bin/git")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(code.INSTALL, path.Base(m.Conf(GIT, kit.META_SOURCE)), arg)
		}},
	},
}

func init() { code.Index.Register(Index, &web.Frame{}) }
