package code

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
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
func init() {
	Index.Register(&ice.Context{Name: "js", Help: "js",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, "js", "js", c.Cap(ice.CTX_FOLLOW))
			}},
			"js": {Name: "js", Help: "js", Action: map[string]*ice.Action{
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
