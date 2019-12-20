package wiki

import (
	"github.com/gomarkdown/markdown"
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"bytes"
	"fmt"
	"path"
	"strings"
)

var prefix = `<svg vertion="1.1" xmlns="http://www.w3.org/2000/svg"
	width="{{.Option "width"}}" height="{{.Option "height"}}" style="{{.Option "style"}}"
	data-name="{{.Option "name"}}"
>`
var title = `<span>{{.Option "prefix"}}{{.Option "content"}}</span>`

var Index = &ice.Context{Name: "wiki", Help: "文档模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"note": {Name: "note", Help: "笔记", Value: map[string]interface{}{
			ice.MDB_META: map[string]interface{}{
				"temp": "var/tmp/file",
				"path": "usr/local/wiki",
				"head": "time size line path",
				"alias": map[string]interface{}{
					"block": []interface{}{"chart", "block"},
					"chain": []interface{}{"chart", "chain"},
					"table": []interface{}{"chart", "table"},

					"chapter": []interface{}{"title", "chapter"},
					"section": []interface{}{"title", "section"},
				},
			},
			ice.MDB_LIST: map[string]interface{}{},
			ice.MDB_HASH: map[string]interface{}{},
		}},
		"chart": {Name: "chart", Help: "绘图", Value: map[string]interface{}{
			ice.MDB_META: map[string]interface{}{
				"prefix": prefix, "suffix": `</svg>`,
			},
			ice.MDB_LIST: map[string]interface{}{},
			ice.MDB_HASH: map[string]interface{}{},
		}},
		"title": {Name: "title", Help: "标题", Value: map[string]interface{}{
			ice.MDB_META: map[string]interface{}{
				"title": title,
			},
			ice.MDB_LIST: map[string]interface{}{},
			ice.MDB_HASH: map[string]interface{}{},
		}},
	},
	Commands: map[string]*ice.Command{
		"chart": {Name: "chart block|chain|table name text [fg bg fs ls p m]", Help: "绘图", Meta: map[string]interface{}{
			"display": "inner",
		}, List: []interface{}{
			map[string]interface{}{"type": "select", "value": "chain", "values": "block chain table"},
			map[string]interface{}{"type": "text", "value": ""},
			map[string]interface{}{"type": "button", "value": "生成"},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 创建类型
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
			arg[2] = strings.TrimSpace(arg[2])

			// 构造数据
			m.Option(ice.WEB_TMPL, "raw")
			m.Option("name", arg[1])
			m.Option("text", arg[2])
			chart.Init(m, arg[2:]...)
			m.Option("width", chart.GetWidth())
			m.Option("height", chart.GetHeight())

			// 生成网页
			m.Render(m.Conf("chart", ice.Meta("prefix")))
			chart.Draw(m, 0, 0)
			m.Render(m.Conf("chart", ice.Meta("suffix")))
			return
		}},

		"_text": {Name: "_text file", Help: "文章", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Option(ice.WEB_TMPL, "raw")
			m.Optionv("title", map[string]int{})

			// 生成文章
			buffer := bytes.NewBuffer([]byte{})
			f := m.Target().Server().(*web.Frame)
			tmpl := f.HandleCGI(m, m.Confm("note", ice.Meta("alias")), path.Join(m.Conf("note", ice.Meta("path")), arg[0]))
			m.Assert(tmpl.ExecuteTemplate(buffer, m.Option("filename", path.Base(arg[0])), m))

			// 缓存文章
			if f, p, e := kit.Create(path.Join(m.Conf("note", ice.Meta("temp")), arg[0])); e == nil {
				defer f.Close()
				if n, e := f.Write(buffer.Bytes()); e == nil {
					m.Log("info", "save %d %v", n, p)
				}
			}

			// 生成网页
			m.Echo(string(markdown.ToHTML(buffer.Bytes(), nil, nil)))
		}},
		"_tree": {Name: "_tree path", Help: "文库", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Cmdy("nfs.dir", m.Conf("note", ice.Meta("path")),
				kit.Select("", arg, 0), m.Conf("note", ice.Meta("head")))
		}},
		"note": {Name: "note file", Help: "笔记", Meta: map[string]interface{}{
			"remote": "true", "display": "inner",
		}, List: []interface{}{
			map[string]interface{}{"type": "text", "value": "miss.md", "name": "path"},
			map[string]interface{}{"type": "button", "value": "执行", "action": "auto"},
			map[string]interface{}{"type": "button", "value": "返回", "cb": "Last"},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Cmdy(kit.Select("_tree", "_text", len(arg) > 0 && strings.HasSuffix(arg[0], ".md")), arg)
		}},

		"title": {Name: "title text", Help: "标题", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			title, _ := m.Optionv("title").(map[string]int)
			switch arg[0] {
			case "section":
				arg = arg[1:]
				title["section"]++
				m.Option("prefix", fmt.Sprintf("%d.%d ", title["chapter"], title["section"]))
			case "chapter":
				arg = arg[1:]
				title["chapter"]++
				title["section"] = 0
				m.Option("prefix", fmt.Sprintf("%d ", title["chapter"]))
			default:
				m.Option("prefix", "")
			}

			ns := strings.Split(m.Conf("runtime", "node.name"), "-")
			m.Option("content", kit.Select(ns[len(ns)-1], arg, 0))
			m.Render(m.Conf("title", ice.Meta("title")))
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
