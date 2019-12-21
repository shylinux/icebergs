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

var prefix = `<svg class="story" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}" vertion="1.1" xmlns="http://www.w3.org/2000/svg"
	width="{{.Option "width"}}" height="{{.Option "height"}}" style="{{.Option "style"}}"
	data-name="{{.Option "name"}}"
>`

var shell = `<div class="story code" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}" data-input="{{.Option "input"}}">$ {{.Option "input"}}
{{.Option "output"}}</div>`

var title = `<span class="story">{{.Option "prefix"}}{{.Option "content"}}</span>`

var order = `<ul class="story" data-name="{{.Option "name"}}" data-text="{{.Option "text"}}">{{range $index, $value := .Optionv "list"}}<li>{{$value}}</li>{{end}}</ul>`

var table = `<table class="story" data-name="{{.Option "name"}} data-text="{{.Option "text"}}>
<tr>{{range $i, $v := .Optionv "head"}}<th>{{$v}}</th>{{end}}</tr>
{{range $index, $value := .Optionv "list"}}
<tr>{{range $i, $v := $value}}<td>{{$v}}</td>{{end}}</tr>
{{end}}
</table>`

var premenu = `<ul class="story premenu"></ul>`
var endmenu = `<ul class="story endmenu">
{{$menu := .Optionv "menu"}}
{{range $index, $value := Value $menu "list"}}
<li>{{Value $value "prefix"}} {{Value $value "content"}}</li>
{{end}}
</ul>
`

var Index = &ice.Context{Name: "wiki", Help: "文档模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"note": {Name: "note", Help: "笔记", Value: ice.Data(
			"temp", "var/tmp/file",
			"path", "usr/local/wiki",
			"head", "time size line path",
			"alias", map[string]interface{}{
				"block": []interface{}{"chart", "block"},
				"chain": []interface{}{"chart", "chain"},

				"chapter": []interface{}{"title", "chapter"},
				"section": []interface{}{"title", "section"},
			},
		)},
		"title": {Name: "title", Help: "标题", Value: ice.Data("template", title)},
		"shell": {Name: "shell", Help: "命令", Value: ice.Data("template", shell)},
		"order": {Name: "order", Help: "列表", Value: ice.Data("template", order)},
		"table": {Name: "table", Help: "表格", Value: ice.Data("template", table)},
		"chart": {Name: "chart", Help: "绘图", Value: ice.Data("prefix", prefix, "suffix", `</svg>`)},
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
			m.Option("name", arg[1])
			m.Option("text", arg[2])
			chart.Init(m, arg[2:]...)
			m.Option("width", chart.GetWidth())
			m.Option("height", chart.GetHeight())

			// 生成网页
			m.Render(m.Conf("chart", ice.Meta("prefix")))
			chart.Draw(m, 0, 0)
			m.Render(m.Conf("chart", ice.Meta("suffix")))
		}},
		"table": {Name: "table name text", Help: "表格", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			head, list := []string{}, [][]string{}
			for i, v := range kit.Split(arg[1], "\n") {
				if i == 0 {
					head = kit.Split(v)
				} else {
					list = append(list, kit.Split(v))
				}
			}
			m.Optionv("head", head)
			m.Optionv("list", list)
			m.Render(m.Conf("table", ice.Meta("template")))
		}},
		"order": {Name: "order name text", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Optionv("name", arg[0])
			m.Optionv("text", arg[1])
			m.Optionv("list", kit.Split(arg[1], "\n"))
			m.Render(m.Conf("order", ice.Meta("template")))
		}},
		"shell": {Name: "shell dir cmd", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Option("cmd_dir", arg[0])
			m.Option("input", strings.Join(arg[1:], " "))
			m.Option("output", m.Cmdx("cli.system", "sh", "-c", strings.Join(arg[1:], " ")))
			m.Render(m.Conf("shell", ice.Meta("template")))
		}},
		"title": {Name: "title text", Help: "标题", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 生成序号
			title, _ := m.Optionv("title").(map[string]int)
			switch arg[0] {
			case "premenu":
				m.Echo(premenu)
				return
			case "endmenu":
				m.Render(endmenu)
				return
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

			// 生成菜单
			menu, _ := m.Optionv("menu").(map[string]interface{})
			list, _ := menu["list"].([]interface{})
			menu["list"] = append(list, map[string]interface{}{
				"content": m.Option("content", kit.Select(ns[len(ns)-1], arg, 0)),
				"prefix":  m.Option("prefix"),
			})

			// 生成网页
			m.Render(m.Conf("title", ice.Meta("template")))
		}},

		"_text": {Name: "_text file", Help: "文章", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Option(ice.WEB_TMPL, "raw")
			m.Optionv("title", map[string]int{})
			m.Optionv("menu", map[string]interface{}{})

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
			"detail": []string{"add", "commit", "favor", "detail"},
		}, List: ice.List(
			ice.MDB_TYPE, "text", "value", "miss.md", "name", "path",
			ice.MDB_TYPE, "button", "value", "执行", "action", "auto",
			ice.MDB_TYPE, "button", "value", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) > 0 {
				switch arg[0] {
				case "story":
					m.Cmdy(arg)
					return
				}
			}
			m.Cmdy(kit.Select("_tree", "_text", len(arg) > 0 && strings.HasSuffix(arg[0], ".md")), arg)
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
