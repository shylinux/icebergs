package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/ssh"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"path"
	"strings"
)

const (
	WORD = "word"

	TITLE = "title"
	CHART = "chart"
	FIELD = "field"
	SHELL = "shell"

	CHAPTER = "chapter"
	SECTION = "section"
	PREMENU = "premenu"
	ENDMENU = "endmenu"

	LABEL = "label"
	CHAIN = "chain"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TITLE: {Name: "title", Help: "标题", Value: kit.Data("template", title)},
			CHART: {Name: "chart", Help: "图表", Value: kit.Data("template", prefix, "suffix", `</svg>`)},
			FIELD: {Name: "field", Help: "插件", Value: kit.Data("template", field,
				"some", kit.Dict("simple", kit.Dict(
					"inputs", kit.List(
						kit.MDB_INPUT, "text", "name", "name",
						kit.MDB_INPUT, "button", "value", "查看", "action", "auto",
						kit.MDB_INPUT, "button", "value", "返回", "cb", "Last",
					),
				)),
			)},
			SHELL: {Name: "shell", Help: "命令", Value: kit.Data("template", shell)},

			WORD: {Name: "word", Help: "语言文字", Value: kit.Data(kit.MDB_SHORT, "name",
				"path", "", "regs", ".*\\.shy", "alias", map[string]interface{}{
					LABEL: []interface{}{CHART, LABEL},
					CHAIN: []interface{}{CHART, CHAIN},

					SECTION: []interface{}{TITLE, SECTION},
					CHAPTER: []interface{}{TITLE, CHAPTER},
					ENDMENU: []interface{}{TITLE, ENDMENU},
					PREMENU: []interface{}{TITLE, PREMENU},
				},
			)},
		},
		Commands: map[string]*ice.Command{
			TITLE: {Name: "title [chapter|section|endmenu|premenu] text", Help: "标题", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				title, _ := m.Optionv(TITLE).(map[string]int)
				switch arg[0] {
				case ENDMENU:
					// 后置目录
					m.Render(ice.RENDER_TEMPLATE, endmenu)
					return
				case PREMENU:
					// 前置目录
					m.Render(ice.RENDER_TEMPLATE, premenu)
					return
				case SECTION:
					// 分节标题
					arg = arg[1:]
					title[SECTION]++
					m.Option("level", "h3")
					m.Option("prefix", fmt.Sprintf("%d.%d", title[CHAPTER], title[SECTION]))
				case CHAPTER:
					// 章节标题
					arg = arg[1:]
					title[CHAPTER]++
					title[SECTION] = 0
					m.Option("level", "h2")
					m.Option("prefix", fmt.Sprintf("%d", title[CHAPTER]))
				default:
					// 文章标题
					m.Option("level", "h1")
					m.Option("prefix", "")
				}

				// 基本参数
				m.Option(kit.MDB_TYPE, TITLE)
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

				// 渲染引擎
				m.Render(ice.RENDER_TEMPLATE, m.Conf(TITLE, "meta.template"))
			}},
			CHART: {Name: "chart label|chain name text", Help: "图表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				var chart Chart
				switch arg[0] {
				case LABEL:
					// 标签
					chart = &Label{}
				case CHAIN:
					// 链接
					chart = &Chain{}
				}
				arg[1] = strings.TrimSpace(arg[1])
				arg[2] = strings.TrimSpace(arg[2])

				// 基本参数
				m.Option(kit.MDB_TYPE, arg[0])
				m.Option(kit.MDB_NAME, arg[1])
				m.Option(kit.MDB_TEXT, arg[2])

				// 扩展参数
				m.Option("font-size", "24")
				m.Option("stroke", "blue")
				m.Option("fill", "yellow")
				// 扩展参数
				m.Option("style", "")
				m.Option("compact", "false")
				m.Option("stroke-width", "2")
				m.Option("padding", "10")
				m.Option("margin", "10")
				// m.Option("font-family", kit.Select("", "monospace", len(arg[2]) == len([]rune(arg[2]))))
				m.Option("font-family", "monospace")
				for i := 3; i < len(arg)-1; i++ {
					m.Option(arg[i], arg[i+1])
				}

				// 计算尺寸
				chart.Init(m, arg[2])
				m.Option("width", chart.GetWidth())
				m.Option("height", chart.GetHeight())

				// 渲染引擎
				m.Render(ice.RENDER_TEMPLATE, m.Conf(CHART, "meta.template"))
				chart.Draw(m, 0, 0)
				m.Render(ice.RENDER_TEMPLATE, m.Conf(CHART, "meta.suffix"))
			}},
			FIELD: {Name: "field name text", Help: "插件", Action: map[string]*ice.Action{
				"run": {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg[1:])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				// 基本参数
				m.Option(kit.MDB_TYPE, FIELD)
				m.Option(kit.MDB_NAME, arg[0])
				m.Option(kit.MDB_TEXT, arg[1])

				// 命令参数
				data := kit.Dict(kit.MDB_NAME, arg[0])
				cmds := kit.Split(arg[1])
				m.Search(cmds[0], func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
					if ls := strings.Split(cmds[0], "."); len(ls) > 1 {
						m.Cmd(ctx.COMMAND, strings.Join(ls[:len(ls)-1], "."), key)
					} else {
						m.Cmd(ctx.COMMAND, key)
					}
					if data["feature"], data["inputs"] = cmd.Meta, cmd.List; len(cmd.List) == 0 {
						data["inputs"] = m.Confv("field", "meta.some.simple.inputs")
					}
				})

				// 扩展参数
				for i := 2; i < len(arg)-1; i += 2 {
					if data := m.Confv("field", kit.Keys("meta.some", arg[i+1], arg[i])); data != nil {
						m.Option(arg[i], data)
					} else {
						m.Parse("option", arg[i], arg[i+1])
					}
					data[arg[i]] = m.Optionv(arg[i])
				}

				// 渲染引擎
				m.Option("meta", data)
				m.Render(ice.RENDER_TEMPLATE, m.Conf(cmd, "meta.template"))
			}},
			SHELL: {Name: "shell name text", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				// 基本参数
				m.Option(kit.MDB_TYPE, SHELL)
				m.Option(kit.MDB_NAME, arg[0])
				m.Option(kit.MDB_TEXT, arg[1])

				// 渲染引擎
				m.Option("input", strings.Join(arg[1:], " "))
				m.Option("output", m.Cmdx(cli.SYSTEM, "sh", "-c", m.Option("input")))
				m.Render(ice.RENDER_TEMPLATE, m.Conf(SHELL, "meta.template"))
			}},

			WORD: {Name: "word path=自然/编程/hi.shy auto", Help: "语言文字", Meta: kit.Dict(
				"display", "/plugin/local/wiki/word.js",
			), Action: map[string]*ice.Action{
				"story": {Name: "story", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg[0], "action", "run", arg[1:])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if reply(m, cmd, arg...) {
					// 目录列表
					return
				}

				// 解析脚本
				m.Option("render", "raw")
				m.Optionv(TITLE, map[string]int{})
				m.Optionv("menu", map[string]interface{}{"list": []interface{}{}})
				m.Optionv(ice.MSG_ALIAS, m.Confv(WORD, "meta.alias"))
				m.Set(ice.MSG_RESULT).Cmdy(ssh.SOURCE, path.Join(m.Conf(WORD, "meta.path"), arg[0]))
			}},
		},
	}, nil)
}
