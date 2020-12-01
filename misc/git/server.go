package git

import (
	"os"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
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
			web.LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(ice.RENDER_OUTPUT, ice.RENDER_RESULT)
			}},
			"/repos/": {Name: "repos", Help: "repos", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(cli.CMD_ENV,
					"GIT_HTTP_EXPORT_ALL", "true",
					"GIT_PROJECT_ROOT", kit.Path("./"),
					"REQUEST_METHOD", m.Option(ice.MSG_METHOD),
					"QUERY_STRING", m.R.URL.RawQuery,
					"PATH", os.Getenv("PATH"),
				)
				m.Cmdy(cli.SYSTEM, "/usr/lib/git-core/git-http-backend")
			}},
		},
	})
}
