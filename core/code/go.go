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

func _go_find(m *ice.Message, key string) {
	for _, p := range strings.Split(m.Cmdx(cli.SYSTEM, "find", ".", "-name", key), "\n") {
		if p == "" {
			continue
		}
		m.Push("file", strings.TrimPrefix(p, "./"))
		m.Push("line", 1)
		m.Push("text", "")
	}
}
func _go_tags(m *ice.Message, key string) {
	if _, e := os.Stat(path.Join(m.Option("_path"), ".tags")); e != nil {
		m.Cmd(cli.SYSTEM, "gotags", "-R", "-f", ".tags", "./")
	}
	for _, l := range strings.Split(m.Cmdx(cli.SYSTEM, "grep", "^"+key, ".tags"), "\n") {
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
				m.Push("file", strings.TrimPrefix(file, "./"))
				m.Push("line", i)
				m.Push("text", bio.Text())
			}
		}
	}
	m.Sort("line", "int")
}
func _go_grep(m *ice.Message, key string) {
	m.Split(m.Cmd(cli.SYSTEM, "grep", "--exclude-dir=.git", "--exclude=.[a-z]*", "-rn", key, ".").Append(cli.CMD_OUT), "file:line:text", ":", "\n")
}
func _go_help(m *ice.Message, key string) {
	p := m.Cmd(cli.SYSTEM, "go", "doc", key).Append(cli.CMD_OUT)
	if p == "" {
		return
	}
	ls := strings.Split(p, "\n")
	if len(ls) > 10 {
		ls = ls[:10]
	}
	res := strings.Join(ls, "\n")

	m.Push("file", key+".godoc")
	m.Push("line", 1)
	m.Push("text", string(res))
}
func init() {
	Index.Register(&ice.Context{Name: "go", Help: "go",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, "go", "go", c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.SEARCH, mdb.CREATE, "godoc", "go", c.Cap(ice.CTX_FOLLOW))
			}},
			"go": {Name: "go", Help: "go", Action: map[string]*ice.Action{
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, m.Option("_path"))
					_go_find(m, kit.Select("main", arg, 1))
					_go_tags(m, kit.Select("main", arg, 1))
					_go_help(m, kit.Select("main", arg, 1))
					_go_grep(m, kit.Select("main", arg, 1))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			}},
		},
	}, nil)
}
