package git

import (
	"os"
	"strings"

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
				m.Option(ice.MSG_USERNAME, "shy")
				m.Option(cli.CMD_ENV,
					"GIT_PROJECT_ROOT", kit.Path("./"),
					"PATH_INFO", "/"+strings.Join(arg, "/"),

					"REMOTE_USER", m.Option(ice.MSG_USERNAME),
					"REMOTE_ADDR", m.Option(ice.MSG_USERNAME),
					"GIT_COMMITTER_NAME", m.Option(ice.MSG_USERNAME),
					"GIT_COMMITTER_EMAIL", m.Option(ice.MSG_USERNAME),

					"REQUEST_METHOD", m.Option(ice.MSG_METHOD),
					"CONTENT_TYPE", m.R.Header.Get(web.ContentType),

					"GIT_HTTP_EXPORT_ALL", "true",
					"QUERY_STRING", m.R.URL.RawQuery,
					"PATH", "/Users/shaoying/miss/contexts/usr/install/git-1.8.3.1"+":"+os.Getenv("PATH"),
				)

				switch strings.Join(arg, "/") {
				case "info/refs":
					msg := m.Cmd(cli.SYSTEM, "/Users/shaoying/miss/contexts/usr/install/git-1.8.3.1"+"/"+"git-http-backend")
					m.Cmd("nfs.file", "append", "hi.log", msg.Append(cli.CMD_ERR))
					x := msg.Result()

					ls := strings.Split(x, "\n")
					for i, v := range ls {
						vs := strings.SplitN(v, ": ", 2)
						if strings.TrimSpace(v) == "" {
							m.Echo(strings.Join(ls[i+1:], "\n") + "\n")
							break
						}
						m.W.Header().Set(vs[0], vs[1])
					}
				case "git-upload-pack":
					m.Option("input", m.R.Body)
					defer m.R.Body.Close()
					msg := m.Cmd(cli.SYSTEM, "/Users/shaoying/miss/contexts/usr/install/git-1.8.3.1"+"/"+"git-upload-pack", "--advertise-refs", kit.Path("./"))
					m.Cmd("nfs.file", "append", "hi.log", msg.Append(cli.CMD_ERR))
					x := msg.Result()

					ls := strings.SplitN(x, "\n", 2)
					m.Debug(" %v %v", len(x), x[:len(x)])

					m.Render(ice.RENDER_OUTPUT, ice.RENDER_VOID)
					m.W.Write([]byte(ls[1]))
				}
			}},
		},
	})
}
