package git

import (
	"github.com/AaronO/go-git-http"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const SERVE = "serve"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			SERVE: {Name: SERVE, Help: "服务", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,name,branch,commit",
				"owner", "https://github.com/shylinux",
			)},
		},
		Commands: map[string]*ice.Command{
			web.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Render(ice.RENDER_RESULT)
			}},
			"/github.com/": {Name: "github.com", Help: "github.com", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Render(ice.RENDER_VOID)
				p := kit.Path(".ish/pluged")
				m.Debug("what %v", p)
				git := githttp.New(p)
				m.Debug("what %v", p)
				git.ServeHTTP(m.W, m.R)
			}},
		},
	})
}
