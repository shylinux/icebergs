package md

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/wiki"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
	"path"
	"strings"

	"github.com/gomarkdown/markdown"
)

/*
func (m *Message) Prefile(favor string, id string) map[string]string {
	// TODO
	res := map[string]string{}
	m.Option("render", "")
	m.Option("_action", "")
	m.Cmd("web.favor", kit.Select(m.Option("favor"), favor), id).Table(func(index int, value map[string]string, head []string) {
		res[value["key"]] = value["value"]
	})

	res["content"] = m.Cmdx("cli.system", "sed", "-n", kit.Format("%d,%dp", kit.Int(res["extra.row"]), kit.Int(res["extra.row"])+3), res["extra.buf"])
	return res
}
*/
var Index = &ice.Context{Name: "md", Help: "md",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"note": {Name: "note", Help: "笔记", Value: kit.Data(
			"path", "", "head", "time size line path",
		)},

		"md": {Name: "md", Help: "md", Value: kit.Data(kit.MDB_SHORT, "name")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {}},

		"md": {Name: "md", Help: "md", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Echo("hello world")
		}},

		"note": {Name: "note file", Help: "文档", Meta: kit.Dict(
			ice.Display("inner"),
		), List: kit.List(
			kit.MDB_INPUT, "text", "name", "path", "value", "README.md", "action", "auto",
			kit.MDB_INPUT, "button", "name", "执行", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && strings.HasSuffix(arg[0], ".md") {
				arg[0] = path.Join(m.Conf("note", "meta.path"), arg[0])
			}
			m.Cmdy(kit.Select("_tree", "_text", len(arg) > 0 && strings.HasSuffix(arg[0], ".md")), arg)
		}},
		"_tree": {Name: "_tree [path [true]]", Help: "文库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("dir_reg", ".*\\.md")
			m.Option("dir_deep", kit.Select("", arg, 1))
			m.Cmdy("nfs.dir", kit.Select(m.Conf("note", "meta.path"), arg, 0), m.Conf("note", "meta.head"))
		}},
		"_text": {Name: "_text file", Help: "文章", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if b, e := ioutil.ReadFile(arg[0]); m.Assert(e) {
				data := markdown.ToHTML(b, nil, nil)
				m.Echo(string(data))
			}
		}},
	},
}

func init() { wiki.Index.Register(Index, nil) }
