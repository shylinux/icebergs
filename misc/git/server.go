package git

import (
	"os"
	"path"
	"strings"

	"github.com/AaronO/go-git-http"
	"github.com/AaronO/go-git-http/auth"
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const SERVER = "server"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVER: {Name: SERVER, Help: "服务器", Value: kit.Data(kit.MDB_PATH, "usr/local")},
		},
		Commands: map[string]*ice.Command{
			SERVER: {Name: "server path auto create", Help: "服务器", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create name", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, path.Join(m.Conf(SERVER, kit.META_PATH), REPOS))
					m.Cmd(cli.SYSTEM, GIT, INIT, "--bare", m.Option(kit.MDB_NAME))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(nfs.DIR_ROOT, path.Join(m.Conf(SERVER, kit.META_PATH), REPOS))
					m.Cmdy(nfs.DIR, "./")
					return
				}

				m.Cmdy("_sum", path.Join("usr/local/repos", arg[0]))
			}},

			web.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Render(ice.RENDER_VOID)
			}},
			"/repos/": {Name: "/github.com/", Help: "/github.com/", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				auth.Authenticator(func(info auth.AuthInfo) (bool, error) {
					if !aaa.UserLogin(m, info.Username, info.Password) {
						return false, nil
					}

					if info.Push {
						if aaa.UserRole(m, info.Username) == aaa.VOID {
							return false, nil
						}

						m.Option(cli.CMD_DIR, path.Join(m.Conf(SERVER, kit.META_PATH), REPOS))
						p := strings.Trim(path.Join(arg...), "info/refs")
						if _, e := os.Stat(path.Join(m.Option(cli.CMD_DIR), p)); os.IsNotExist(e) {
							m.Cmd(cli.SYSTEM, GIT, INIT, "--bare", p)
						}
					}

					return true, nil
				})(githttp.New(kit.Path(m.Conf(SERVER, kit.META_PATH)))).ServeHTTP(m.W, m.R)
			}},
			"/github.com/": {Name: "/github.com/", Help: "/github.com/", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				auth.Authenticator(func(info auth.AuthInfo) (bool, error) {
					if !aaa.UserLogin(m, info.Username, info.Password) {
						return false, nil
					}

					if info.Push && aaa.UserRole(m, info.Username) == aaa.VOID {
						return false, nil
					}

					return true, nil
				})(githttp.New(kit.Path(m.Conf(SERVER, kit.META_PATH)))).ServeHTTP(m.W, m.R)
			}},
		},
	})
}
