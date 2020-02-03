package wiki

import (
	"github.com/gomarkdown/markdown"
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"bytes"
	"fmt"
	"os"
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
				"label": []interface{}{"chart", "block"},
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

		"mind": {Name: "mind", Help: "思维导图", Value: kit.Data(kit.MDB_SHORT, "name", "path", "usr/local", "regs", ".*\\.svg", "prefix", `<svg vertion="1.1" xmlns="http://www.w3.org/2000/svg" width="%v" height="%v">`, "suffix", `</svg>`)},
		"word": {Name: "word", Help: "语言文字", Value: kit.Data(kit.MDB_SHORT, "name", "path", "usr/local", "regs", ".*\\.shy")},
		"data": {Name: "data", Help: "数据表格", Value: kit.Data(kit.MDB_SHORT, "name", "path", "usr/local", "regs", ".*\\.csv")},
		"feel": {Name: "feel", Help: "影音媒体", Value: kit.Data(kit.MDB_SHORT, "name", "path", "usr/local", "regs", ".*\\.(JPG|MOV|m4v)")},
		"walk": {Name: "walk", Help: "走遍世界", Value: kit.Data(kit.MDB_SHORT, "name", "path", "usr/local", "regs", ".*\\.csv")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", kit.Keys(m.Cap(ice.CTX_FOLLOW), "json"))

		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", kit.Keys(m.Cap(ice.CTX_FOLLOW), "json"), kit.Keys(m.Cap(ice.CTX_FOLLOW), "feel"))
		}},
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
				chart = &Table{}
			case "chain":
				chart = &Chain{}
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
		"stack": {Name: "stack name text", Help: "堆栈", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			chain := &Chain{}
			m.Render(m.Conf("spark", ice.Meta("template")), arg[0])
			stack(m, "stack", 0, kit.Parse(nil, "", chain.show(m, arg[1])...))
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

			input, output := "", ""
			switch arg = arg[2:]; arg[0] {
			case "install", "compile":
				input = strings.Join(arg[1:], " ")
			default:
				input = strings.Join(arg, " ")
				output = m.Cmdx("cli.system", "sh", "-c", strings.Join(arg, " "))
			}

			m.Option("input", input)
			m.Option("output", output)
			m.Render(m.Conf("spark", ice.Meta("template")), m.Option("name"))
			m.Render(m.Conf("shell", ice.Meta("template")))
		}},
		"index": {Name: "index hash", Help: "索引", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Cmd(ice.WEB_STORY, "index", arg)
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
			list := [][]string{}
			for _, v := range kit.Split(arg[1], "\n") {
				list = append(list, kit.Split(v, " "))
			}
			m.Optionv("list", list)
			m.Render(m.Conf("refer", ice.Meta("template")))
		}},
		"brief": {Name: "brief name text", Help: "摘要", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
			// if strings.HasSuffix(arg[0], ".md") {
			data = markdown.ToHTML(buffer.Bytes(), nil, nil)
			// }
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

		"mind": {Name: "mind", Help: "思维导图", Meta: kit.Dict("display", "wiki/mind"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "button", "name", "执行",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "保存":
					m.Cmd("nfs.save", path.Join(m.Conf(cmd, "meta.path"), arg[2]), arg[3:])
				}
				return
			}

			// 文件列表
			m.Option("dir_root", m.Conf(cmd, "meta.path"))
			m.Option("dir_reg", m.Conf(cmd, "meta.regs"))
			m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
			m.Sort("time", "time_r")
			if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
				// 目录列表
				m.Option("dir_reg", "")
				m.Option("dir_type", "dir")
				m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
			}
		}},
		"word": {Name: "word", Help: "语言文字", Meta: kit.Dict("display", "wiki/word"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "button", "name", "执行",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "保存":
					m.Cmd("nfs.save", path.Join(m.Conf(cmd, "meta.path"), arg[2]), arg[3])
				}
				return
			}

			// 文件列表
			m.Option("dir_root", m.Conf(cmd, "meta.path"))
			m.Option("dir_reg", m.Conf(cmd, "meta.regs"))
			m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
			m.Sort("time", "time_r")
			if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
				// 目录列表
				m.Option("dir_reg", "")
				m.Option("dir_type", "dir")
				m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
				return
			}
			m.Option("preview", m.Cmdx("_text", path.Join(m.Conf(cmd, "meta.path"), arg[0])))
		}},
		"data": {Name: "data", Help: "数据表格", Meta: kit.Dict("display", "wiki/data"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "button", "name", "执行",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "保存":
					m.Cmd("nfs.save", path.Join(m.Conf("mind", "meta.path"), arg[2]), arg[3])
				}
				return
			}

			// 文件列表
			m.Option("dir_root", m.Conf(cmd, "meta.path"))
			m.Option("dir_reg", m.Conf(cmd, "meta.regs"))
			m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
			m.Sort("time", "time_r")
			if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
				// 目录列表
				m.Option("dir_reg", "")
				m.Option("dir_type", "dir")
				m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
				return
			}
			m.CSV(m.Result())
		}},
		"feel": {Name: "feel", Help: "影音媒体", Meta: kit.Dict("display", "wiki/feel"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "button", "name", "执行",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "保存":
					m.Cmd("nfs.save", path.Join(m.Conf("mind", "meta.path"), arg[2]), arg[3])
				case "标签":
					m.Conf("feel", kit.Keys(path.Base(arg[2]), "-2"), arg[3])
					p := path.Join(m.Conf(cmd, "meta.path"), arg[2])
					q := path.Join(m.Conf(cmd, "meta.path"), arg[3])
					os.MkdirAll(q, 0777)
					m.Assert(os.Link(p, path.Join(q, path.Base(arg[2]))))
				}
				return
			}

			if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
				// 文件列表
				m.Option("dir_root", m.Conf(cmd, "meta.path"))
				m.Option("dir_reg", m.Conf(cmd, "meta.regs"))
				m.Cmdy("nfs.dir", kit.Select("./", arg, 0), "time size path").Table(func(index int, value map[string]string, head []string) {
					m.Push("label", m.Conf("feel", path.Base(value["path"])))
				})

				// 目录列表
				m.Option("dir_reg", "")
				m.Option("dir_type", "dir")
				m.Cmdy("nfs.dir", kit.Select("./", arg, 0))

				if len(arg) > 0 {
					m.Sort("time", "time_r")
				} else {
					m.Sort("line", "int_r")
				}
				return
			}
			// 下载文件
			m.Echo(path.Join(m.Conf(cmd, "meta.path"), arg[0]))
		}},
		"walk": {Name: "walk", Help: "走遍世界", Meta: kit.Dict("display", "wiki/walk"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "file",
			kit.MDB_INPUT, "button", "name", "执行",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "保存":
					m.Cmd("nfs.save", path.Join(m.Conf("mind", "meta.path"), arg[2]), arg[3])
				}
				return
			}

			// 文件列表
			m.Option("dir_root", m.Conf(cmd, "meta.path"))
			m.Option("dir_reg", m.Conf(cmd, "meta.regs"))
			m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
			m.Sort("time", "time_r")
			if len(arg) == 0 || strings.HasSuffix(arg[0], "/") {
				// 目录列表
				m.Option("dir_reg", "")
				m.Option("dir_type", "dir")
				m.Cmdy("nfs.dir", kit.Select("./", arg, 0))
				return
			}
			m.Option("title", "我走过的世界")
			m.CSV(m.Result())
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
