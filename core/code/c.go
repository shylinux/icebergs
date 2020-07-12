package code

import (
	"bufio"
	"os"
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _c_find(m *ice.Message, key string) {
	m.Option(cli.CMD_DIR, "")
	for _, p := range strings.Split(m.Cmdx(cli.SYSTEM, "find", m.Option("_path"), "-name", key), "\n") {
		if p == "" {
			continue
		}
		m.Push("file", p)
		m.Push("line", 1)
		m.Push("text", "")
	}
}
func _c_tags(m *ice.Message, key string) {
	m.Option(cli.CMD_DIR, m.Option("_path"))
	if _, e := os.Stat(path.Join(m.Option("_path"), ".tags")); e != nil {
		m.Cmd(cli.SYSTEM, "ctags", "-R", "-f", ".tags", "./")
	}
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, "grep", key, ".tags"), "\n") {
		ls := strings.SplitN(l, "\t", 2)
		if len(ls) < 2 {
			continue
		}

		ls = strings.SplitN(ls[1], "\t", 2)
		file := ls[0]
		ls = strings.SplitN(ls[1], ";\"", 2)
		text := strings.TrimSuffix(strings.TrimPrefix(ls[0], "/^"), "$/")
		line := kit.Int(text)

		p := path.Join(m.Option("_path"), file)
		f, e := os.Open(p)
		m.Assert(e)
		bio := bufio.NewScanner(f)
		for i := 1; bio.Scan(); i++ {
			if i == line || bio.Text() == text {
				m.Push("file", p)
				m.Push("line", i)
				m.Push("text", bio.Text())
			}
		}
	}
}
func _c_grep(m *ice.Message, key string) {
	m.Option(cli.CMD_DIR, "")
	m.Split(m.Cmdx(cli.SYSTEM, "grep", "--exclude=.*", "-rn", key, m.Option("_path")), "file:line:text", ":", "\n")
}
func init() {
	Index.Register(&ice.Context{Name: "c", Help: "c",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, "h", "c", c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.SEARCH, mdb.CREATE, "c", "c", c.Cap(ice.CTX_FOLLOW))
			}},
			"c": {Name: "c", Help: "c", Action: map[string]*ice.Action{
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					_c_find(m, kit.Select("main", arg, 1))
					_c_tags(m, kit.Select("main", arg, 1))
					_c_grep(m, kit.Select("main", arg, 1))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			}},
		},
	}, nil)
}
