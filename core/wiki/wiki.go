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
			"path", "", "temp", "var/tmp/file",
			"head", "time size line path",
			"alias", map[string]interface{}{
				"label": []interface{}{"chart", "label"},
				"chain": []interface{}{"chart", "chain"},

				"section": []interface{}{"title", "section"},
				"chapter": []interface{}{"title", "chapter"},
				"endmenu": []interface{}{"title", "endmenu"},
				"premenu": []interface{}{"title", "premenu"},
			},
		)},

		"title": {Name: "title", Help: "标题", Value: kit.Data("template", title)},
		"brief": {Name: "brief", Help: "摘要", Value: kit.Data("template", brief)},
		"refer": {Name: "refer", Help: "参考", Value: kit.Data("template", refer)},
		"spark": {Name: "spark", Help: "段落", Value: kit.Data("template", spark)},

		"local": {Name: "local", Help: "文件", Value: kit.Data("template", local)},
		"shell": {Name: "shell", Help: "命令", Value: kit.Data("template", shell)},
		"order": {Name: "order", Help: "列表", Value: kit.Data("template", order)},
		"table": {Name: "table", Help: "表格", Value: kit.Data("template", table)},
		"stack": {Name: "stack", Help: "结构", Value: kit.Data("template", stack)},
		"chart": {Name: "chart", Help: "绘图", Value: kit.Data("prefix", prefix, "suffix", `</svg>`)},

		"draw": {Name: "draw", Help: "思维导图", Value: kit.Data(kit.MDB_SHORT, "name", "path", "usr/local", "regs", ".*\\.svg", "prefix", `<svg vertion="1.1" xmlns="http://www.w3.org/2000/svg" width="%v" height="%v">`, "suffix", `</svg>`)},
		"word": {Name: "word", Help: "语言文字", Value: kit.Data(kit.MDB_SHORT, "name", "path", "usr/local", "regs", ".*\\.shy")},
		"data": {Name: "data", Help: "数据表格", Value: kit.Data(kit.MDB_SHORT, "name", "path", "usr/local", "regs", ".*\\.csv")},
		"feel": {Name: "feel", Help: "影音媒体", Value: kit.Data(kit.MDB_SHORT, "name", "path", "usr/local", "regs", ".*\\.(png|JPG|MOV|m4v)")},
		"walk": {Name: "walk", Help: "走遍世界", Value: kit.Data(kit.MDB_SHORT, "name", "path", "usr/local", "regs", ".*\\.csv")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("feel")
		}},

		"note": {Name: "note file", Help: "笔记", Meta: kit.Dict("remote", "you", "display", "inner"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "path", "value", "README.md",
			kit.MDB_INPUT, "button", "name", "执行", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
		"_tree": {Name: "_tree path", Help: "文库", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// m.Option("dir_deep", "true")
			m.Option("dir_reg", ".*\\.md")
			m.Cmdy("nfs.dir", kit.Select(m.Conf("note", "meta.path"), arg, 0), m.Conf("note", "meta.head"))
		}},
		"_text": {Name: "_text file", Help: "文章", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(ice.WEB_TMPL, "raw")
			m.Optionv("title", map[string]int{})
			m.Optionv("menu", map[string]interface{}{"list": []interface{}{}})
			if strings.HasSuffix(arg[0], ".shy") {
				m.Optionv(ice.MSG_ALIAS, m.Confv("note", "meta.alias"))
				m.Cmdy("ssh.scan", arg[0], arg[0], arg[0])
				return
			}

			// 生成文章
			buffer := bytes.NewBuffer([]byte{})
			f := m.Target().Server().(*web.Frame)
			tmpl := f.HandleCGI(m, m.Confm("note", "meta.alias"), arg[0])
			m.Assert(tmpl.ExecuteTemplate(buffer, m.Option("filename", path.Base(arg[0])), m))

			// 缓存文章
			if f, p, e := kit.Create(path.Join(m.Conf("note", "meta.temp"), arg[0])); e == nil {
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

		"title": {Name: "title [chapter|section|endmenu|premenu] text", Help: "标题", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			title, _ := m.Optionv("title").(map[string]int)
			switch arg[0] {
			case "endmenu":
				// 后置目录
				m.Render(endmenu)
				return
			case "premenu":
				// 前置目录
				m.Render(premenu)
				return
			case "section":
				arg = arg[1:]
				title["section"]++
				m.Option("level", "h3")
				m.Option("prefix", fmt.Sprintf("%d.%d", title["chapter"], title["section"]))
			case "chapter":
				arg = arg[1:]
				title["chapter"]++
				title["section"] = 0
				m.Option("level", "h2")
				m.Option("prefix", fmt.Sprintf("%d", title["chapter"]))
			default:
				m.Option("level", "h1")
				m.Option("prefix", "")
			}
			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option(kit.MDB_TEXT, arg[0])

			// 添加目录
			ns := strings.Split(m.Conf("runtime", "node.name"), "-")
			menu, _ := m.Optionv("menu").(map[string]interface{})
			menu["list"] = append(menu["list"].([]interface{}), map[string]interface{}{
				"content": m.Option("content", kit.Select(ns[len(ns)-1], arg, 0)),
				"prefix":  m.Option("prefix"),
				"level":   m.Option("level"),
			})

			// 生成网页
			m.Render(m.Conf("title", "meta.template"))
		}},
		"brief": {Name: "brief name text", Help: "摘要", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Echo(`<br class="story" data-type="brief">`)
				return
			}
			if len(arg) == 1 {
				arg = []string{"", arg[0]}
			}
			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option(kit.MDB_TEXT, arg[1])
			m.Render(m.Conf(cmd, "meta.template"))
		}},
		"refer": {Name: "refer name text", Help: "参考", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option(kit.MDB_TEXT, arg[1])

			list := [][]string{}
			for _, v := range kit.Split(strings.TrimSpace(arg[1]), "\n") {
				list = append(list, kit.Split(v, " "))
			}
			m.Optionv("list", list)
			m.Render(m.Conf(cmd, "meta.template"))
		}},
		"spark": {Name: "spark name text", Help: "感悟", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.Echo(`<br class="story" data-type="spark">`)
				return
			}
			if len(arg) == 1 {
				arg = []string{"", arg[0]}
			}

			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option(kit.MDB_TEXT, arg[1])
			m.Optionv("list", kit.Split(arg[1], "\n"))
			m.Render(m.Conf(cmd, "meta.template"))
		}},

		"local": {Name: "local name text", Help: "文件", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option(kit.MDB_TEXT, arg[1])
			m.Option("input", m.Cmdx("nfs.cat", arg[1]))

			switch ls := strings.Split(arg[1], "."); ls[len(ls)-1] {
			case "csv":
				list := []string{"<table>"}
				m.Spawn().CSV(m.Option("input")).Table(func(index int, value map[string]string, head []string) {
					if index == 0 {
						list = append(list, "<tr>")
						for _, k := range head {
							list = append(list, "<th>", k, "</th>")
						}
						list = append(list, "</tr>")
					}

					list = append(list, "<tr>")
					for _, k := range head {
						list = append(list, "<td>", value[k], "</td>")
					}
					list = append(list, "</tr>")
				})
				list = append(list, "</table>")
				m.Optionv("input", list)
			}
			m.Render(m.Conf(cmd, "meta.template"))
		}},
		"shell": {Name: "shell name dir cmd", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option("cmd_dir", arg[1])

			input, output := "", ""
			switch arg = arg[2:]; arg[0] {
			case "install", "compile":
				input = strings.Join(arg[1:], " ")
			default:
				input = strings.Join(arg, " ")
				output = m.Cmdx(ice.CLI_SYSTEM, "sh", "-c", input)
			}

			m.Option("input", input)
			m.Option("output", output)
			m.Render(m.Conf(cmd, "meta.template"))
		}},
		"order": {Name: "order name text", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option(kit.MDB_TEXT, arg[1])
			m.Optionv("list", kit.Split(strings.TrimSpace(arg[1]), "\n"))
			m.Render(m.Conf(cmd, "meta.template"))
		}},
		"table": {Name: "table name text", Help: "表格", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option(kit.MDB_TEXT, arg[1])

			head, list := []string{}, [][]string{}
			for i, v := range kit.Split(strings.TrimSpace(arg[1]), "\n") {
				if i == 0 {
					head = kit.Split(v)
				} else {
					line := kit.Split(v)
					for i, v := range line {
						if ls := kit.Split(v); len(ls) > 1 {
							style := []string{}
							for i := 1; i < len(ls)-1; i += 2 {
								switch ls[i] {
								case "bg":
									ls[i] = "background-color"
								case "fg":
									ls[i] = "color"
								}
								style = append(style, ls[i]+":"+ls[i+1])
							}
							line[i] = kit.Format(`<span style="%s">%s</span>`, strings.Join(style, ";"), ls[0])
						}
					}
					list = append(list, line)
				}
			}
			m.Optionv("head", head)
			m.Optionv("list", list)
			m.Render(m.Conf(cmd, "meta.template"))
		}},
		"stack": {Name: "stack name text", Help: "结构", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option(kit.MDB_TEXT, arg[1])

			chain := &Chain{}
			m.Render(m.Conf(cmd, "meta.template"))
			Stack(m, cmd, 0, kit.Parse(nil, "", chain.show(m, arg[1])...))
			m.Echo("</div>")
		}},
		"chart": {Name: "chart label|chain|table name text [fg bg fs ls p m]", Help: "绘图", Meta: map[string]interface{}{}, List: kit.List(
			kit.MDB_INPUT, "select", "value", "chain", "values", "block chain table",
			kit.MDB_INPUT, "text", "value", "",
			kit.MDB_INPUT, "button", "value", "生成",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			// 创建类型
			var chart Chart
			switch arg[0] {
			case "label":
				chart = &Label{}
			case "chain":
				chart = &Chain{}
			}
			arg[1] = strings.TrimSpace(arg[1])
			arg[2] = strings.TrimSpace(arg[2])

			// 构造数据
			m.Option(kit.MDB_TYPE, arg[0])
			m.Option(kit.MDB_NAME, arg[1])
			m.Option(kit.MDB_TEXT, arg[2])
			m.Option("font-size", kit.Select("16", arg, 3))
			m.Option("stroke", kit.Select("yellow", arg, 4))
			m.Option("fill", kit.Select("purple", arg, 5))

			m.Option("style", "")
			m.Option("compact", "false")
			m.Option("stroke-width", "2")
			m.Option("font-family", kit.Select("", "monospace", len(arg[2]) == len([]rune(arg[2]))))
			for i := 6; i < len(arg)-1; i++ {
				m.Option(arg[i], arg[i+1])
			}

			chart.Init(m, arg[2:]...)
			m.Option("width", chart.GetWidth())
			m.Option("height", chart.GetHeight())

			// 生成网页
			m.Render(m.Conf("chart", "meta.prefix"))
			chart.Draw(m, 4, 4)
			m.Render(m.Conf("chart", "meta.suffix"))
		}},

		"draw": {Name: "draw", Help: "思维导图", Meta: kit.Dict("display", "wiki/draw"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "name", "value", "what/he.svg",
			kit.MDB_INPUT, "button", "name", "执行", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
			kit.MDB_INPUT, "button", "name", "上传", "cb", "upload",
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
		"word": {Name: "word", Help: "语言文字", Meta: kit.Dict("remote", "pod", "display", "wiki/word"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "name", "value", "自然/编程/hi.shy",
			kit.MDB_INPUT, "button", "name", "执行", "action", "auto",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
			kit.MDB_INPUT, "button", "name", "上传", "cb", "upload",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Option("_action") == "上传" {
				if len(arg) == 0 {
					arg = append(arg, "/")
				}
				m.Cmdy(ice.WEB_STORY, "upload")
				m.Cmd(ice.WEB_STORY, ice.STORY_WATCH, m.Append("data"),
					path.Join(m.Conf(cmd, "meta.path"), arg[0], kit.Select("", m.Append("name"), strings.HasSuffix(arg[0], "/"))))
				return
			}

			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "追加":
					if f, e := os.OpenFile(path.Join(m.Conf(cmd, "meta.path"), arg[2]), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666); m.Assert(e) {
						defer f.Close()
						f.WriteString("\n")
						f.WriteString(arg[3])

						if len(arg) > 4 {
							f.WriteString(` "`)
							f.WriteString(arg[4])
							f.WriteString(`"`)
						}
						if len(arg) > 5 {
							f.WriteString(" `")
							f.WriteString(arg[5])
							f.WriteString("`")
						}
						for _, v := range arg[6:] {
							f.WriteString(" `")
							f.WriteString(v)
							f.WriteString("`")
						}

						f.WriteString("\n")
						f.WriteString("\n")
					}

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

			// 解析脚本
			m.Option(ice.WEB_TMPL, "raw")
			m.Optionv("title", map[string]int{})
			m.Optionv("menu", map[string]interface{}{"list": []interface{}{}})
			m.Optionv(ice.MSG_ALIAS, m.Confv("note", "meta.alias"))
			m.Set("result").Cmdy("ssh.scan", arg[0], arg[0], path.Join(m.Conf(cmd, "meta.path"), arg[0]))
		}},
		"data": {Name: "data", Help: "数据表格", Meta: kit.Dict("display", "wiki/data"), List: kit.List(
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "button", "name", "执行",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
			kit.MDB_INPUT, "button", "name", "上传", "cb", "upload",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Option("_action") == "上传" {
				if len(arg) == 0 {
					arg = append(arg, "/")
				}
				m.Cmdy(ice.WEB_STORY, "upload")
				m.Cmd(ice.WEB_STORY, ice.STORY_WATCH, m.Append("data"),
					path.Join(m.Conf(cmd, "meta.path"), arg[0], kit.Select("", m.Append("name"), strings.HasSuffix(arg[0], "/"))))
				return
			}

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
			m.CSV(m.Result())
		}},
		"feel": {Name: "feel", Help: "影音媒体", Meta: kit.Dict("display", "wiki/feel", "detail", []string{"标签", "删除"}), List: kit.List(
			kit.MDB_INPUT, "text", "name", "name",
			kit.MDB_INPUT, "button", "name", "执行",
			kit.MDB_INPUT, "button", "name", "返回", "cb", "Last",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "删除":
					m.Assert(os.Remove(path.Join(m.Conf(cmd, "meta.path"), m.Option("path"))))
				case "保存":
					m.Cmd("nfs.save", path.Join(m.Conf(cmd, "meta.path"), arg[2]), arg[3])
				case "标签":
					m.Conf(cmd, kit.Keys(path.Base(arg[2]), "-2"), arg[3])
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
					m.Push("label", m.Conf(cmd, path.Base(value["path"])))
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
			m.Option("title", "我走过的世界")
			m.CSV(m.Result())
		}},

		"mind": {Name: "mind zone type name text", Help: "思考", List: kit.List(
			kit.MDB_INPUT, "text", "name", "path", "action", "auto", "figure", "key",
			kit.MDB_INPUT, "text", "name", "type", "figure", "key",
			kit.MDB_INPUT, "text", "name", "name", "figure", "key",
			kit.MDB_INPUT, "button", "name", "添加",
			kit.MDB_INPUT, "textarea", "name", "text",
			kit.MDB_INPUT, "text", "name", "location", "figure", "key", "cb", "location",
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "input":
					// 输入补全
					switch arg[2] {
					case "type":
						m.Push("type", []string{"spark", "label", "refer", "brief", "chapter", "section", "title"})
					case "path":
						m.Option("_refresh", "true")
						// 文件列表
						m.Option("dir_root", m.Conf("word", "meta.path"))
						m.Option("dir_reg", m.Conf("word", "meta.regs"))
						m.Cmdy("nfs.dir", kit.Select("./", arg, 3), "path")
						m.Sort("time", "time_r")
						if len(arg) == 3 || strings.HasSuffix(arg[3], "/") {
							// 目录列表
							m.Option("dir_reg", "")
							m.Option("dir_type", "dir")
							m.Cmdy("nfs.dir", kit.Select("./", arg, 3), "path")
							return
						}
					}
					return
				}
			}

			if len(arg) < 2 {
				m.Cmdy("word", arg)
				return
			}

			m.Cmd("word", "action", "追加", arg)
			m.Option("scan_mode", "scan")
			m.Cmdy("ssh.scan", "some", "some", path.Join(m.Conf("word", "meta.path"), arg[0]))
		}},

		"qrcode": {Name: "qrcode", Help: "扫码", Meta: kit.Dict("display", "wiki/image"), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			data := map[string]interface{}{}
			for i := 0; i < len(arg)-1; i += 2 {
				kit.Value(data, arg[i], arg[i+1])
			}
			m.Push("_output", "qrcode")
			m.Echo(kit.Format(data))
		}},
		"qrcode2": {Name: "qrcode2", Help: "扫码", Meta: kit.Dict("display", "wiki/image"), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Push("_output", "qrcode")
			m.Echo(kit.MergeURL(arg[0], arg[1:]))
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
