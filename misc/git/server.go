package git

import (
	"github.com/AaronO/go-git-http"
	"github.com/AaronO/go-git-http/auth"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const SERVER = "server"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVER: {Name: SERVER, Help: "服务器", Value: kit.Data(kit.MDB_PATH, ".ish/pluged")},
		},
		Commands: map[string]*ice.Command{
			web.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Render(ice.RENDER_VOID)
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
