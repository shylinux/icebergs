package wiki

import (
	"bytes"
	"github.com/gomarkdown/markdown"
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
	"path"
	"strings"
)

var Index = &ice.Context{Name: "wiki", Help: "文档模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"note": {Name: "note", Value: map[string]interface{}{
			"meta": map[string]interface{}{
				"path": "usr/local/wiki",
			},
			"list": map[string]interface{}{},
			"hash": map[string]interface{}{},
		}},
	},
	Commands: map[string]*ice.Command{
		"chart": {Name: "chart", Help: "绘图", List: []interface{}{
			map[string]interface{}{"type": "select", "value": "chain", "values": "chain table"},
			map[string]interface{}{"type": "text", "value": ""},
			map[string]interface{}{"type": "button", "value": "执行"},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("render", "raw")
			var chart Chart
			switch arg[0] {
			case "block":
				chart = &Block{}
			case "chain":
				chart = &Chain{}
			case "table":
				chart = &Table{}
			}
			arg[1] = strings.TrimSpace(arg[1])

			chart.Init(m, arg[1:]...)
			m.Echo(`<svg vertion="1.1" xmlns="http://www.w3.org/2000/svg" width="%d", height="%d" style="%s">`,
				chart.GetWidth(), chart.GetHeight(), m.Option("style"))
			m.Echo("\n")
			chart.Draw(m, 0, 0)
			m.Echo(`</svg>`)
			m.Echo("\n")
			return
		}},

		"_tree": {Name: "_tree", Help: "目录", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Cmdy("nfs.dir", m.Conf("note", "meta.path"), kit.Select("", arg, 0), "time size line path")
		}},
		"_text": {Name: "_text", Help: "文章", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			tmpl := m.Target().Server().(*web.WEB).HandleCGI(m, path.Join(m.Conf("note", "meta.path"), arg[0]))
			m.Optionv("title", map[string]int{})

			buffer := bytes.NewBuffer([]byte{})
			m.Assert(tmpl.ExecuteTemplate(buffer, m.Option("filename", path.Base(arg[0])), m))

			if f, p, e := kit.Create(path.Join("var/tmp/file", arg[0])); e == nil {
				defer f.Close()
				if n, e := f.Write(buffer.Bytes()); e == nil {
					m.Log("info", "save %d %v", n, p)
				}
			}

			data := markdown.ToHTML(buffer.Bytes(), nil, nil)
			m.Echo(string(data))
		}},
		"note": {Name: "note file|favor|commit", Help: "笔记", Meta: map[string]interface{}{
			"display": "inner",
			"remote":  "true",
		}, List: []interface{}{
			map[string]interface{}{"type": "text", "value": "miss.md", "name": "path"},
			map[string]interface{}{"type": "button", "value": "执行", "action": "auto"},
			map[string]interface{}{"type": "button", "value": "返回", "cb": "Last"},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy("_tree")
				return
			}

			switch arg[0] {
			case "favor", "commit":
				m.Cmdy("story", arg[0], arg[1:])
			default:
				m.Cmdy(kit.Select("_tree", "_text", strings.HasSuffix(arg[0], ".md")), arg[0])
			}
		}},

		"title": {Name: "title text", Help: "一级标题", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			ns := strings.Split(m.Conf("runtime", "node.name"), "-")
			m.Option("render", cmd)
			m.Echo(kit.Select(ns[len(ns)-1], arg, 0))
		}},
	},
}

func init() { web.Index.Register(Index, &web.WEB{}) }
