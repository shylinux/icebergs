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
	for _, p := range strings.Split(m.Cmdx(cli.SYSTEM, "find", ".", "-name", key), "\n") {
		if p == "" {
			continue
		}
		m.Push("file", strings.TrimPrefix(p, "./"))
		m.Push("line", 1)
		m.Push("text", "")
	}
}
func _c_tags(m *ice.Message, key string) {
	if _, e := os.Stat(path.Join(m.Option("_path"), ".tags")); e != nil {
		m.Cmd(cli.SYSTEM, "ctags", "-R", "-f", ".tags", "./")
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
func _c_grep(m *ice.Message, key string) {
	m.Split(m.Cmd(cli.SYSTEM, "grep",
		"--exclude-dir=.git", "--exclude-dir=pluged",
		"--exclude=.[a-z]*", "-rn", key, ".").Append(cli.CMD_OUT), "file:line:text", ":", "\n")
}
func _c_help(m *ice.Message, section, key string) {
	p := m.Cmd(cli.SYSTEM, "man", section, key).Append(cli.CMD_OUT)
	if p == "" {
		return
	}
	ls := strings.Split(p, "\n")

	if len(ls) > 20 {
		p = strings.Join(ls[:20], "\n")
	}
	p = strings.Replace(p, "_\x08", "", -1)
	res := make([]byte, 0, len(p))
	for i := 0; i < len(p); i++ {
		switch p[i] {
		case '\x08':
			i++
		default:
			res = append(res, p[i])
		}
	}

	m.Push("file", key+".man"+section)
	m.Push("line", 1)
	m.Push("text", string(res))
}

const C = "c"
const H = "h"
const MAN1 = "man1"
const MAN2 = "man2"
const MAN3 = "man3"
const MAN8 = "man8"

func init() {
	Index.Register(&ice.Context{Name: C, Help: "c",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, H, C, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.SEARCH, mdb.CREATE, C, C, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.SEARCH, mdb.CREATE, MAN3, C, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.SEARCH, mdb.CREATE, MAN2, C, c.Cap(ice.CTX_FOLLOW))
			}},
			C: {Name: C, Help: "c", Action: map[string]*ice.Action{
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, m.Option("_path"))
					_c_find(m, kit.Select("main", arg, 1))
					_c_help(m, "2", kit.Select("main", arg, 1))
					_c_help(m, "3", kit.Select("main", arg, 1))
					_c_tags(m, kit.Select("main", arg, 1))
					_c_grep(m, kit.Select("main", arg, 1))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			}},
		},
	}, nil)
}
