package wiki

import (
	ice "github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"os"
	"path"
	"strings"
)

func reply(m *ice.Message, cmd string, arg ...string) bool {
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
		m.Option("_display", "table")
		return true
	}
	return false
}

var Index = &ice.Context{Name: "wiki", Help: "文档中心",
	Configs: map[string]*ice.Config{
		"brief": {Name: "brief", Help: "摘要", Value: kit.Data("template", brief)},
		"refer": {Name: "refer", Help: "参考", Value: kit.Data("template", refer)},
		"spark": {Name: "spark", Help: "段落", Value: kit.Data("template", spark)},

		"local": {Name: "local", Help: "文件", Value: kit.Data("template", local)},

		"order": {Name: "order", Help: "列表", Value: kit.Data("template", order)},
		"table": {Name: "table", Help: "表格", Value: kit.Data("template", table)},
		"stack": {Name: "stack", Help: "结构", Value: kit.Data("template", stack)},

		"data": {Name: "data", Help: "数据表格", Value: kit.Data(kit.MDB_SHORT, "name", "path", "", "regs", ".*\\.csv")},
		"walk": {Name: "walk", Help: "走遍世界", Value: kit.Data(kit.MDB_SHORT, "name", "path", "", "regs", ".*\\.csv")},
		"feel": {Name: "feel", Help: "影音媒体", Value: kit.Data(kit.MDB_SHORT, "name", "path", "", "regs", ".*\\.(png|jpg|JPG|MOV|m4v)")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("feel")
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

			m.Render(ice.RENDER_TEMPLATE, m.Conf(cmd, "meta.template"))
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
			m.Render(ice.RENDER_TEMPLATE, m.Conf(cmd, "meta.template"))
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
			m.Render(ice.RENDER_TEMPLATE, m.Conf(cmd, "meta.template"))
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
			m.Render(ice.RENDER_TEMPLATE, m.Conf(cmd, "meta.template"))
		}},

		"order": {Name: "order name text", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option(kit.MDB_TEXT, arg[1])

			m.Optionv("list", kit.Split(strings.TrimSpace(arg[1]), "\n"))
			m.Render(ice.RENDER_TEMPLATE, m.Conf(cmd, "meta.template"))
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
			m.Render(ice.RENDER_TEMPLATE, m.Conf(cmd, "meta.template"))
		}},
		"stack": {Name: "stack name text", Help: "结构", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(kit.MDB_TYPE, cmd)
			m.Option(kit.MDB_NAME, arg[0])
			m.Option(kit.MDB_TEXT, arg[1])

			chain := &Chain{}
			m.Render(ice.RENDER_TEMPLATE, m.Conf(cmd, "meta.template"))
			Stack(m, cmd, 0, kit.Parse(nil, "", chain.show(m, arg[1])...))
			m.Echo("</div>")
		}},

		"data": {Name: "data path auto", Help: "数据表格", Meta: kit.Dict("display", "local/wiki/data"), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) > 0 && arg[0] == "action" {
				switch arg[1] {
				case "保存":
					m.Cmd("nfs.save", path.Join(m.Conf(cmd, "meta.path"), arg[2]), arg[3])
				}
				return
			}

			if reply(m, cmd, arg...) {
				// 目录列表
				return
			}
			// 解析数据
			m.CSV(m.Result())
		}},
		"feel": {Name: "feel path auto 上传:button=@upload", Help: "影音媒体", Meta: kit.Dict(
			"display", "local/wiki/feel", "detail", []string{"标签", "删除"},
		), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Option("_action") == "上传" {
				m.Cmd(ice.WEB_CACHE, "watch", m.Option("_data"), path.Join(m.Option("name"), m.Option("_name")))
				return
			}

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
		"walk": {Name: "walk path=@province auto", Help: "走遍世界", Meta: kit.Dict("display", "local/wiki/walk"), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
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
						m.Push("type", []string{"spark", "order", "table", "label", "chain", "refer", "brief", "chapter", "section", "title"})
					case "path":
						m.Option("_refresh", "true")
						reply(m, "word", arg[3:]...)
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
			m.Cmdy(ice.SSH_SOURCE, path.Join(m.Conf("word", "meta.path"), arg[0]))
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
