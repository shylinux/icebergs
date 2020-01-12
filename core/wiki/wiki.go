package wiki

import (
	"github.com/gomarkdown/markdown"
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"bytes"
	"fmt"
	"html"
	"path"
	"strings"
)

var Index = &ice.Context{Name: "wiki", Help: "文档中心",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"note": {Name: "note", Help: "笔记", Value: kit.Data(
			"temp", "var/tmp/file",
			"path", "",
			"head", "time size line path",
			"alias", map[string]interface{}{
				"block": []interface{}{"chart", "block"},
				"chain": []interface{}{"chart", "chain"},

				"chapter": []interface{}{"title", "chapter"},
				"section": []interface{}{"title", "section"},
			},
		)},
		"title": {Name: "title", Help: "标题", Value: kit.Data("template", title)},
		"brief": {Name: "brief", Help: "摘要", Value: kit.Data("template", brief)},
		"refer": {Name: "refer", Help: "参考", Value: kit.Data("template", refer)},
		"spark": {Name: "spark", Help: "段落", Value: kit.Data("template", spark)},

		"shell": {Name: "shell", Help: "命令", Value: kit.Data("template", shell)},
		"order": {Name: "order", Help: "列表", Value: kit.Data("template", order)},
		"table": {Name: "table", Help: "表格", Value: kit.Data("template", table)},
		"chart": {Name: "chart", Help: "绘图", Value: kit.Data("prefix", prefix, "suffix", `</svg>`)},
	},
	Commands: map[string]*ice.Command{
		"chart": {Name: "chart block|chain|table name text [fg bg fs ls p m]", Help: "绘图", Meta: map[string]interface{}{
			"display": "inner",
		}, List: kit.List(
			kit.MDB_INPUT, "select", "value", "chain", "values", "block chain table",
			kit.MDB_INPUT, "text", "value", "",
			kit.MDB_INPUT, "button", "value", "生成",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
			m.Option("type", arg[0])
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
		"table": {Name: "table name text", Help: "表格", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("type", "table")
			m.Option("name", arg[0])
			m.Option("text", arg[1])
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
			m.Option("type", "order")
			m.Option("name", arg[0])
			m.Option("text", arg[1])
			m.Optionv("list", kit.Split(arg[1], "\n"))
			m.Render(m.Conf("order", ice.Meta("template")))
		}},
		"shell": {Name: "shell name dir cmd", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Option("type", "shell")
			m.Option("name", arg[0])
			m.Option("cmd_dir", arg[1])

			switch arg = arg[2:]; arg[0] {
			case "install", "compile":
				m.Option("input", html.EscapeString(strings.Join(arg[1:], " ")))
			default:
				m.Option("input", html.EscapeString(strings.Join(arg, " ")))
				m.Option("output", html.EscapeString(m.Cmdx("cli.system", "sh", "-c", strings.Join(arg, " "))))
			}
			m.Render(m.Conf("spark", ice.Meta("template")), m.Option("name"))
			m.Render(m.Conf("shell", ice.Meta("template")))
		}},

		"spark": {Name: "spark name text", Help: "参考", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Option("type", "refer")
			m.Option("name", arg[0])
			m.Option("text", arg[1])
			m.Optionv("list", kit.Split(arg[1], "\n"))
			m.Render(m.Conf("order", ice.Meta("template")))
		}},
		"refer": {Name: "refer name text", Help: "参考", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Option("type", "refer")
			m.Option("name", arg[0])
			m.Option("text", arg[1])
			list := []interface{}{}
			for _, v := range kit.Split(arg[1], "\n") {
				list = append(list, kit.Split(v, " "))
			}
			m.Optionv("list", list)
			m.Render(m.Conf("order", ice.Meta("template")))
		}},
		"brief": {Name: "brief text", Help: "摘要", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option("type", "brief")
			m.Option("name", arg[0])
			m.Option("text", arg[1])
			m.Render(m.Conf("brief", ice.Meta("template")))
		}},
		"title": {Name: "title text", Help: "标题", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 生成序号
			m.Option("level", "h1")
			title, _ := m.Optionv("title").(map[string]int)
			switch arg[0] {
			case "endmenu":
				m.Render(endmenu)
				return
			case "premenu":
				m.Render(premenu)
				return
			case "section":
				m.Option("level", "h3")
				arg = arg[1:]
				title["section"]++
				m.Option("prefix", fmt.Sprintf("%d.%d ", title["chapter"], title["section"]))
			case "chapter":
				m.Option("level", "h2")
				arg = arg[1:]
				title["chapter"]++
				title["section"] = 0
				m.Option("prefix", fmt.Sprintf("%d ", title["chapter"]))
			default:
				m.Option("prefix", "")
			}
			m.Option("type", "title")
			m.Option("text", arg[0])

			// 生成菜单
			ns := strings.Split(m.Conf("runtime", "node.name"), "-")
			menu, _ := m.Optionv("menu").(map[string]interface{})
			menu["list"] = append(menu["list"].([]interface{}), map[string]interface{}{
				"content": m.Option("content", kit.Select(ns[len(ns)-1], arg, 0)),
				"prefix":  m.Option("prefix"),
			})

			// 生成网页
			m.Render(m.Conf("title", ice.Meta("template")))
		}},

		"_text": {Name: "_text file", Help: "文章", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Option(ice.WEB_TMPL, "raw")
			m.Optionv("title", map[string]int{})
			m.Optionv("menu", map[string]interface{}{"list": []interface{}{}})

			// 生成文章
			buffer := bytes.NewBuffer([]byte{})
			f := m.Target().Server().(*web.Frame)
			tmpl := f.HandleCGI(m, m.Confm("note", ice.Meta("alias")), arg[0])
			m.Assert(tmpl.ExecuteTemplate(buffer, m.Option("filename", path.Base(arg[0])), m))

			// 缓存文章
			if f, p, e := kit.Create(path.Join(m.Conf("note", ice.Meta("temp")), arg[0])); e == nil {
				defer f.Close()
				if n, e := f.Write(buffer.Bytes()); e == nil {
					m.Log("info", "save %d %v", n, p)
				}
			}

			// 生成网页
			data := buffer.Bytes()
			if strings.HasSuffix(arg[0], ".md") {
				data = markdown.ToHTML(buffer.Bytes(), nil, nil)
			}
			m.Echo(string(data))
		}},
		"_tree": {Name: "_tree path", Help: "文库", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Option("dir_deep", "true")
			m.Option("dir_reg", ".*\\.md")
			m.Cmdy("nfs.dir", kit.Select(m.Conf("note", "meta.path"), arg, 0), m.Conf("note", "meta.head"))
		}},
		"note": {Name: "note file", Help: "笔记", Meta: kit.Dict("remote", "you", "display", "inner"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "path", "value", "README.md",
			kit.MDB_INPUT, "button", "name", "执行", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) > 1 {
				switch arg[1] {
				case "运行":
					switch arg[2] {
					case "shell":
						m.Cmdy(ice.CLI_SYSTEM, "sh", "-c", arg[4])
					}

				case "favor":
					m.Cmdy(ice.WEB_FAVOR, kit.Select("story", m.Option("hot")), arg[2:])
				case "share":
					m.Cmdy(ice.WEB_SHARE, "add", arg[2:])
				default:
					m.Cmdy(arg)
				}
				return
			}
			if len(arg) > 0 && strings.HasSuffix(arg[0], ".md") {
				arg[0] = path.Join(m.Conf("note", "meta.path"), arg[0])
			}
			m.Cmdy(kit.Select("_tree", "_text", len(arg) > 0 && strings.HasSuffix(arg[0], ".md")), arg)
		}},
		"index": {Name: "index hash", Help: "索引", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Cmd(ice.WEB_STORY, "index", arg)
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
