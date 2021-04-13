package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/ssh"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"net/url"
	"path"
	"strings"
)

func _name(m *ice.Message, arg []string) []string {
	if len(arg) == 1 {
		return []string{"", arg[0]}
	}
	return arg
}
func _option(m *ice.Message, kind, name, text string, arg ...string) {
	m.Option(kit.MDB_TYPE, kind)
	m.Option(kit.MDB_NAME, name)
	m.Option(kit.MDB_TEXT, text)

	extra := kit.Dict()
	m.Optionv(kit.MDB_EXTRA, extra)
	for i := 0; i < len(arg); i += 2 {
		extra[arg[i]] = kit.Format(kit.Parse(nil, "", kit.Split(arg[i+1])...))
	}
}

func _title_show(m *ice.Message, kind, text string, arg ...string) {
	title, _ := m.Optionv(TITLE).(map[string]int)
	switch kind {
	case PREMENU:
		// 前置目录
		m.Render(ice.RENDER_TEMPLATE, premenu)
		return
	case ENDMENU:
		// 后置目录
		m.Render(ice.RENDER_TEMPLATE, endmenu)
		return
	case SECTION:
		// 分节标题
		title[SECTION]++
		m.Option("level", "h3")
		m.Option("prefix", fmt.Sprintf("%d.%d ", title[CHAPTER], title[SECTION]))
	case CHAPTER:
		// 章节标题
		title[CHAPTER]++
		title[SECTION] = 0
		m.Option("level", "h2")
		m.Option("prefix", fmt.Sprintf("%d ", title[CHAPTER]))
	default:
		// 文章标题
		m.Option("level", "h1")
		m.Option("prefix", "")
	}

	// 添加目录
	menu, _ := m.Optionv("menu").(map[string]interface{})
	menu["list"] = append(menu["list"].([]interface{}), map[string]interface{}{
		"content": m.Option("content", text),
		"prefix":  m.Option("prefix"),
		"level":   m.Option("level"),
	})

	// 渲染引擎
	_option(m, TITLE, text, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(TITLE, "meta.template"))
}
func _brief_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, BRIEF, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(BRIEF, "meta.template"))
}
func _refer_show(m *ice.Message, name, text string, arg ...string) {
	list := [][]string{}
	for _, v := range kit.Split(strings.TrimSpace(text), "\n", "\n") {
		if ls := kit.Split(v); len(ls) == 1 {
			list = append(list, []string{path.Base(ls[0]), ls[0]})
		} else {
			list = append(list, ls)
		}
	}
	m.Optionv("list", list)

	_option(m, REFER, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(REFER, "meta.template"))
}
func _spark_show(m *ice.Message, name, text string, arg ...string) {
	switch text = strings.TrimSpace(text); name {
	case "shell", "redis", "mysql":
		m.Echo(`<div class="story" data-type="spark" data-name="%s">`, name)
		for _, l := range strings.Split(text, "\n") {
			m.Echo("<div>")
			m.Echo("<label>")
			m.Echo(kit.Select(name+"> ", m.Conf(SPARK, kit.Keys("meta.prompt", name))))
			m.Echo("</label>").Echo("<span>").Echo(l).Echo("</span>")
			m.Echo("</div>")
		}
		m.Echo("</div>")
		return
	}
	m.Optionv("list", kit.Split(text, "\n", "\n"))

	_option(m, SPARK, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(SPARK, "meta.template"))
}

func _order_show(m *ice.Message, name, text string, arg ...string) {
	m.Optionv("list", kit.Split(strings.TrimSpace(text), "\n"))
	_option(m, ORDER, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(ORDER, "meta.template"))
}
func _table_show(m *ice.Message, name, text string, arg ...string) {
	head, list := []string{}, [][]string{}
	for i, v := range kit.Split(strings.TrimSpace(text), "\n") {
		if v = strings.ReplaceAll(v, "%", "%%"); i == 0 {
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

	_option(m, TABLE, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(TABLE, "meta.template"))
}
func _shell_show(m *ice.Message, name, text string, arg ...string) {
	m.Option("output", m.Cmdx(cli.SYSTEM, "sh", "-c", m.Option("input", text)))
	_option(m, SHELL, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(SHELL, "meta.template"))
}
func _local_show(m *ice.Message, name, text string, arg ...string) {
	m.Option("input", m.Cmdx(nfs.CAT, text))
	_option(m, LOCAL, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(LOCAL, "meta.template"))
}

func _image_show(m *ice.Message, name, text string, arg ...string) {
	if name == "qrcode" {
		m.EchoQRCode(text)
		return
	}
	if !strings.HasPrefix(text, "http") && !strings.HasPrefix(text, "/") {
		text = "/share/local/usr/local/image/" + text
	}

	_option(m, IMAGE, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(IMAGE, "meta.template"))
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
	m.Option("marginx", "10")
	m.Option("marginy", "10")
	// m.Option("font-family", kit.Select("", "monospace", len(text) == len([]rune(text))))
	m.Option("font-family", "monospace")
	for i := 0; i < len(arg)-1; i++ {
		m.Option(arg[i], arg[i+1])
	}
	if m.Option("fg") != "" {
		m.Option("stroke", m.Option("fg"))
	}
	if m.Option("bg") != "" {
		m.Option("fill", m.Option("bg"))
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
		data["feature"], data["inputs"] = cmd.Meta, cmd.List
	})

	// 扩展参数
	for i := 0; i < len(arg)-1; i += 2 {
		if strings.HasPrefix(arg[i], "args.") {
			m.Option(arg[i], strings.TrimSpace(arg[i+1]))
			kit.Value(data, arg[i], m.Option(arg[i]))
		} else if strings.HasPrefix(arg[i], "args") {
			m.Option(arg[i], kit.Split(strings.TrimSuffix(strings.TrimPrefix(arg[i+1], "["), "]")))
			kit.Value(data, arg[i], m.Optionv(arg[i]))
		} else {
			m.Parse("option", arg[i], arg[i+1])
			kit.Value(data, arg[i], m.Optionv(arg[i]))
		}

		switch arg[i] {
		case "content":
			data[arg[i]] = arg[i+1]

		case "args":
			args := kit.Simple(m.Optionv(arg[i]))

			count := 0
			kit.Fetch(data["inputs"], func(index int, value map[string]interface{}) {
				if value["_input"] != "button" && value["type"] != "button" {
					count++
				}
			})

			if len(args) > count {
				list := data["inputs"].([]interface{})
				for i := count; i < len(args); i++ {
					list = append(list, kit.Dict(
						"_input", "text", "name", "args", "value", args[i],
					))
				}
				data["inputs"] = list
			}
		}
	}

	// 渲染引擎
	m.Option("meta", data)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(FIELD, "meta.template"))
}
func _other_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, OTHER, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(OTHER, "meta.template"))
}

func _word_show(m *ice.Message, name string, arg ...string) {
	m.Set(ice.MSG_RESULT)
	m.Option("render", "raw")
	m.Option(TITLE, map[string]int{})
	m.Option("menu", map[string]interface{}{"list": []interface{}{}})

	m.Option(ice.MSG_ALIAS, m.Confv(WORD, "meta.alias"))
	m.Option(nfs.DIR_ROOT, _wiki_path(m, WORD))
	m.Cmdy(ssh.SOURCE, name)
}

const (
	TITLE = "title"
	BRIEF = "brief"
	REFER = "refer"
	SPARK = "spark"

	ORDER = "order"
	TABLE = "table"
	SHELL = "shell"
	LOCAL = "local"

	IMAGE = "image"
	CHART = "chart"
	FIELD = "field"
	OTHER = "other"

	PARSE = "parse"

	PREMENU = "premenu"
	CHAPTER = "chapter"
	SECTION = "section"
	ENDMENU = "endmenu"

	LABEL = "label"
	CHAIN = "chain"
)

const WORD = "word"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			TITLE: {Name: TITLE, Help: "标题", Value: kit.Data("template", title)},
			BRIEF: {Name: BRIEF, Help: "摘要", Value: kit.Data("template", brief)},
			REFER: {Name: REFER, Help: "参考", Value: kit.Data("template", refer)},
			SPARK: {Name: SPARK, Help: "段落", Value: kit.Data("template", spark, "prompt", kit.Dict("shell", "$ "))},

			ORDER: {Name: ORDER, Help: "列表", Value: kit.Data("template", order)},
			TABLE: {Name: TABLE, Help: "表格", Value: kit.Data("template", table)},
			SHELL: {Name: SHELL, Help: "命令", Value: kit.Data("template", shell)},
			LOCAL: {Name: LOCAL, Help: "文件", Value: kit.Data("template", local)},

			IMAGE: {Name: IMAGE, Help: "图片", Value: kit.Data("template", image)},
			CHART: {Name: CHART, Help: "图表", Value: kit.Data("template", chart, "suffix", `</svg>`)},
			FIELD: {Name: FIELD, Help: "插件", Value: kit.Data("template", field)},
			OTHER: {Name: FIELD, Help: "网页", Value: kit.Data("template", other)},

			WORD: {Name: WORD, Help: "语言文字", Value: kit.Data(
				kit.MDB_PATH, "", "regs", ".*\\.shy", "alias", map[string]interface{}{
					PREMENU: []interface{}{TITLE, PREMENU},
					CHAPTER: []interface{}{TITLE, CHAPTER},
					SECTION: []interface{}{TITLE, SECTION},
					ENDMENU: []interface{}{TITLE, ENDMENU},
					LABEL:   []interface{}{CHART, LABEL},
					CHAIN:   []interface{}{CHART, CHAIN},
				},
			)},
		},
		Commands: map[string]*ice.Command{
			TITLE: {Name: "title [premenu|chapter|section|endmenu] text", Help: "标题", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					ns := strings.Split(ice.Info.NodeName, "-")
					arg = append(arg, kit.Select(ns[len(ns)-1], ""))
				}

				if len(arg) == 1 {
					arg = append(arg, "")
				}
				_title_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			BRIEF: {Name: "brief [name] text", Help: "摘要", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_brief_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			REFER: {Name: "refer [name] `[name url]...`", Help: "参考", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_refer_show(m, arg[0], arg[1], arg[2:]...)
			}},
			SPARK: {Name: "spark [name] text", Help: "段落", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Echo(`<br class="story" data-type="spark">`)
					return
				}

				arg = _name(m, arg)
				_spark_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},

			ORDER: {Name: "order [name] `[item \n]...`", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_order_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			TABLE: {Name: "table [name] `[item item\n]...`", Help: "表格", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_table_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			SHELL: {Name: "shell [name] cmd", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_shell_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			LOCAL: {Name: "local [name] file", Help: "文件", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_local_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},

			IMAGE: {Name: "image [name] url", Help: "图片", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_image_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
				m.Render("")
			}},
			CHART: {Name: "chart label|chain [name] text", Help: "图表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 2 {
					arg = []string{arg[0], "", arg[1]}
				}
				_chart_show(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			FIELD: {Name: "field [name] cmd", Help: "插件", Action: map[string]*ice.Action{
				"run": {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					if !m.Warn(!m.Right(arg[1:]), ice.ErrNotRight, arg[1:]) {
						m.Cmdy(arg[1:])
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_field_show(m, strings.ReplaceAll(kit.Select(path.Base(arg[1]), arg[0]), " ", "_"), arg[1], arg[2:]...)
			}},
			OTHER: {Name: "other [name] url", Help: "网页", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_other_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},

			PARSE: {Name: "parse type=auto,json,http,form,list auto text:textarea", Help: "解析", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) < 2 {
					return
				}
				if arg[0] == "auto" && (strings.HasPrefix(arg[1], "{") || strings.HasPrefix(arg[1], "[")) {
					arg[0] = "json"
				} else if strings.HasPrefix(arg[1], "http") {
					arg[0] = "http"
				} else if strings.Contains(arg[1], "=") {
					arg[0] = "form"
				} else {
					arg[0] = "list"
				}

				m.Option(mdb.FIELDS, mdb.DETAIL)
				switch arg[0] {
				case "json":
					m.Echo(kit.Formats(kit.UnMarshal(arg[1])))
				case "http":
					u, _ := url.Parse(arg[1])
					for k, v := range u.Query() {
						for _, v := range v {
							m.Push(k, v)
						}
					}
					m.EchoQRCode(arg[1])

				case "form":
					for _, v := range kit.Split(arg[1], "&", "&", "&") {
						ls := kit.Split(v, "=", "=", "=")
						key, _ := url.QueryUnescape(ls[0])
						value, _ := url.QueryUnescape(kit.Select("", ls, 1))
						m.Push(key, value)
					}
				case "list":
					for i, v := range kit.Split(arg[1]) {
						m.Push(kit.Format(i), v)
					}
				}
			}},

			WORD: {Name: "word path=src/main.shy auto 演示", Help: "语言文字", Meta: kit.Dict(
				kit.MDB_DISPLAY, "/plugin/local/wiki/word.js", kit.MDB_STYLE, WORD,
			), Action: map[string]*ice.Action{
				web.STORY: {Name: "story", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg[0], kit.MDB_ACTION, "run", arg[1:])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(nfs.DIR_DEEP, "true"); _wiki_list(m, cmd, arg...) {
					return
				}
				_word_show(m, arg[0])
			}},
		},
	})
}
