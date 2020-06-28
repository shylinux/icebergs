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

func _title_show(m *ice.Message, kind, text string, arg ...string) {
	title, _ := m.Optionv(TITLE).(map[string]int)
	switch kind {
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
		title[SECTION]++
		m.Option("level", "h3")
		m.Option("prefix", fmt.Sprintf("%d.%d", title[CHAPTER], title[SECTION]))
	case CHAPTER:
		// 章节标题
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
	m.Option(kit.MDB_NAME, text)
	m.Option(kit.MDB_TEXT, text)

	// 添加目录
	menu, _ := m.Optionv("menu").(map[string]interface{})
	menu["list"] = append(menu["list"].([]interface{}), map[string]interface{}{
		"content": m.Option("content", text),
		"prefix":  m.Option("prefix"),
		"level":   m.Option("level"),
	})

	// 渲染引擎
	m.Render(ice.RENDER_TEMPLATE, m.Conf(TITLE, "meta.template"))
}
func _brief_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(kit.MDB_TYPE, BRIEF)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	// 渲染引擎
	m.Render(ice.RENDER_TEMPLATE, m.Conf(BRIEF, "meta.template"))
}
func _refer_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(kit.MDB_TYPE, REFER)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	list := [][]string{}
	for _, v := range kit.Split(strings.TrimSpace(text), "\n") {
		list = append(list, kit.Split(v, " "))
	}
	m.Optionv("list", list)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(REFER, "meta.template"))
}
func _spark_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(kit.MDB_TYPE, SPARK)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	m.Optionv("list", kit.Split(text, "\n"))
	m.Render(ice.RENDER_TEMPLATE, m.Conf(SPARK, "meta.template"))
}

func _chart_show(m *ice.Message, kind, name, text string, arg ...string) {
	var chart Chart
	switch kind {
	case LABEL:
		// 标签
		chart = &Label{}
	case CHAIN:
		// 链接
		chart = &Chain{}
	}
	name = strings.TrimSpace(name)
	text = strings.TrimSpace(text)

	// 基本参数
	m.Option(kit.MDB_TYPE, kind)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

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
	// m.Option("font-family", kit.Select("", "monospace", len(text) == len([]rune(text))))
	m.Option("font-family", "monospace")
	for i := 0; i < len(arg)-1; i++ {
		m.Option(arg[i], arg[i+1])
	}

	// 计算尺寸
	chart.Init(m, text)
	m.Option("width", chart.GetWidth())
	m.Option("height", chart.GetHeight())

	// 渲染引擎
	m.Render(ice.RENDER_TEMPLATE, m.Conf(CHART, "meta.template"))
	chart.Draw(m, 0, 0)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(CHART, "meta.suffix"))
}
func _field_show(m *ice.Message, name, text string, arg ...string) {
	// 基本参数
	m.Option(kit.MDB_TYPE, FIELD)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	// 命令参数
	data := kit.Dict(kit.MDB_NAME, name)
	cmds := kit.Split(text)
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
	for i := 0; i < len(arg)-1; i += 2 {
		if data := m.Confv("field", kit.Keys("meta.some", arg[i+1], arg[i])); data != nil {
			m.Option(arg[i], data)
		} else {
			m.Parse("option", arg[i], arg[i+1])
		}
		data[arg[i]] = m.Optionv(arg[i])
	}

	// 渲染引擎
	m.Option("meta", data)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(FIELD, "meta.template"))
}
func _shell_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(kit.MDB_TYPE, SHELL)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	// 渲染引擎
	m.Option("output", m.Cmdx(cli.SYSTEM, "sh", "-c", m.Option("input", text)))
	m.Render(ice.RENDER_TEMPLATE, m.Conf(SHELL, "meta.template"))
}
func _local_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(kit.MDB_TYPE, LOCAL)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	m.Option("input", m.Cmdx("nfs.cat", text))
	m.Render(ice.RENDER_TEMPLATE, m.Conf(LOCAL, "meta.template"))
}

func _order_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(kit.MDB_TYPE, ORDER)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	m.Optionv("list", kit.Split(strings.TrimSpace(text), "\n"))
	m.Render(ice.RENDER_TEMPLATE, m.Conf(ORDER, "meta.template"))
}
func _table_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(kit.MDB_TYPE, TABLE)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	head, list := []string{}, [][]string{}
	for i, v := range kit.Split(strings.TrimSpace(text), "\n") {
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
	m.Render(ice.RENDER_TEMPLATE, m.Conf(TABLE, "meta.template"))
}
func _stack_show(m *ice.Message, name, text string, arg ...string) {
	m.Option(kit.MDB_TYPE, STACK)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	chain := &Chain{}
	m.Render(ice.RENDER_TEMPLATE, m.Conf(STACK, "meta.template"))
	Stack(m, STACK, 0, kit.Parse(nil, "", chain.show(m, text)...))
	m.Echo("</div>")
}

func _word_show(m *ice.Message, name string, arg ...string) {
	m.Set(ice.MSG_RESULT)
	m.Option("render", "raw")
	m.Optionv(TITLE, map[string]int{})
	m.Optionv("menu", map[string]interface{}{"list": []interface{}{}})
	m.Optionv(ice.MSG_ALIAS, m.Confv(WORD, "meta.alias"))
	m.Cmdy(ssh.SOURCE, path.Join(m.Conf(WORD, "meta.path"), name))
}

const (
	WORD = "word"

	TITLE = "title"
	BRIEF = "brief"
	REFER = "refer"
	SPARK = "spark"

	CHART = "chart"
	FIELD = "field"
	SHELL = "shell"
	LOCAL = "local"

	ORDER = "order"
	TABLE = "table"
	STACK = "stack"

	CHAPTER = "chapter"
	SECTION = "section"
	ENDMENU = "endmenu"
	PREMENU = "premenu"

	LABEL = "label"
	CHAIN = "chain"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TITLE: {Name: "title", Help: "标题", Value: kit.Data("template", title)},
			BRIEF: {Name: "brief", Help: "摘要", Value: kit.Data("template", brief)},
			REFER: {Name: "refer", Help: "参考", Value: kit.Data("template", refer)},
			SPARK: {Name: "spark", Help: "段落", Value: kit.Data("template", spark)},

			CHART: {Name: "chart", Help: "图表", Value: kit.Data("template", chart, "suffix", `</svg>`)},
			FIELD: {Name: "field", Help: "插件", Value: kit.Data("template", field)},
			SHELL: {Name: "shell", Help: "命令", Value: kit.Data("template", shell)},
			LOCAL: {Name: "local", Help: "文件", Value: kit.Data("template", local)},

			ORDER: {Name: "order", Help: "列表", Value: kit.Data("template", order)},
			TABLE: {Name: "table", Help: "表格", Value: kit.Data("template", table)},
			STACK: {Name: "stack", Help: "结构", Value: kit.Data("template", stack)},

			WORD: {Name: "word", Help: "语言文字", Value: kit.Data(kit.MDB_SHORT, "name",
				"path", "usr/demo", "regs", ".*\\.shy", "alias", map[string]interface{}{
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
				if len(arg) == 0 {
					ns := strings.Split(cli.NodeName, "-")
					arg = append(arg, kit.Select(ns[len(ns)-1], ""))
				}
				if len(arg) == 1 {
					arg = append(arg, arg[0])
				}
				_title_show(m, arg[0], arg[1], arg[2:]...)
			}},
			BRIEF: {Name: "brief name text", Help: "摘要", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Echo(`<br class="story" data-type="brief">`)
					return
				}
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_brief_show(m, arg[0], arg[1], arg[2:]...)
			}},
			REFER: {Name: "refer name `[name url]...`", Help: "参考", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_refer_show(m, arg[0], arg[1], arg[2:]...)
			}},
			SPARK: {Name: "spark name text", Help: "感悟", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Echo(`<br class="story" data-type="spark">`)
					return
				}
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_spark_show(m, arg[0], arg[1], arg[2:]...)
			}},

			CHART: {Name: "chart label|chain name text arg...", Help: "图表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_chart_show(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			FIELD: {Name: "field name cmd", Help: "插件", Action: map[string]*ice.Action{
				"run": {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg[1:])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_field_show(m, arg[0], arg[1], arg[2:]...)
			}},
			SHELL: {Name: "shell name cmd", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_shell_show(m, arg[0], arg[1])
			}},
			LOCAL: {Name: "local name text", Help: "文件", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_local_show(m, arg[0], arg[1])
			}},

			ORDER: {Name: "order name text", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_order_show(m, arg[0], arg[1], arg[2:]...)
			}},
			TABLE: {Name: "table name text", Help: "表格", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_table_show(m, arg[0], arg[1], arg[2:]...)
			}},
			STACK: {Name: "stack name text", Help: "结构", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_stack_show(m, arg[0], arg[1], arg[2:]...)
			}},

			WORD: {Name: "word path=hi.shy auto", Help: "语言文字", Meta: kit.Dict(
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
				_word_show(m, arg[0])
			}},
		},
	}, nil)
}
