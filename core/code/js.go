package code

import (
	"net/http"
	"os"
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

func _js_find(m *ice.Message, key string) {
	for _, p := range strings.Split(m.Cmdx(cli.SYSTEM, "find", ".", "-name", key), "\n") {
		if p == "" {
			continue
		}
		m.Push("file", strings.TrimPrefix(p, "./"))
		m.Push("line", 1)
		m.Push("text", "")
	}
}
func _js_grep(m *ice.Message, key string) {
	m.Split(m.Cmd(cli.SYSTEM, "grep", "--exclude-dir=.git", "--exclude=.[a-z]*", "-rn", key, ".").Append(cli.CMD_OUT), "file:line:text", ":", "\n")
}

const JS = "js"
const NODE = "node"

func init() {
	Index.Register(&ice.Context{Name: JS, Help: "js",
		Configs: map[string]*ice.Config{
			NODE: {Name: NODE, Help: "服务器", Value: kit.Data(
				"source", "https://nodejs.org/dist/v10.13.0/node-v10.13.0-linux-x64.tar.xz",
			)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, JS, JS, c.Cap(ice.CTX_FOLLOW))
			}},
			NODE: {Name: NODE, Help: "node", Action: map[string]*ice.Action{
				"install": {Name: "install", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					// 下载
					source := m.Conf(NODE, "meta.source")
					p := path.Join(m.Conf("web.code._install", "meta.path"), path.Base(source))
					if _, e := os.Stat(p); e != nil {
						msg := m.Cmd(web.SPIDE, "dev", web.CACHE, http.MethodGet, source)
						m.Cmd(web.CACHE, web.WATCH, msg.Append(web.DATA), p)
					}

					// 解压
					m.Option(cli.CMD_DIR, m.Conf("web.code._install", "meta.path"))
					m.Cmd(cli.SYSTEM, "tar", "xvf", path.Base(source))
					m.Echo(p)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			}},
			JS: {Name: JS, Help: "js", Action: map[string]*ice.Action{
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, m.Option("_path"))
					_js_find(m, kit.Select("main", arg, 1))
					_js_grep(m, kit.Select("main", arg, 1))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			}},
		},
	}, nil)
}
