package wiki

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/ssh"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
	"github.com/skip2/go-qrcode"

	"bytes"
	"encoding/base64"
	"fmt"
	"path"
	"strings"
)

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
	for _, v := range kit.Split(strings.TrimSpace(text), "\n") {
		list = append(list, kit.Split(v, " "))
	}
	m.Optionv("list", list)

	_option(m, REFER, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(REFER, "meta.template"))
}
func _spark_show(m *ice.Message, name, text string, arg ...string) {
	text = strings.TrimSpace(text)
	m.Optionv("list", kit.Split(text, "\n"))

	_option(m, SPARK, name, text, arg...)
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
		if arg[i] == "args" {
			args := kit.Simple(m.Optionv(arg[i]))

			count := 0
			kit.Fetch(data["inputs"], func(index int, value map[string]interface{}) {
				if value["_input"] == "text" || value["type"] == "text" {
					count++
				}
			})

			if len(args) > count {
				list := data["inputs"].([]interface{})
				for i := count; i < len(args); i++ {
					list = append(list, kit.Dict(
						"_input", "text",
						"name", "args",
						"value", args[i],
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
func _image_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, IMAGE, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(IMAGE, "meta.template"))
}
func _video_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, VIDEO, name, text, arg...)
	m.Render(ice.RENDER_TEMPLATE, m.Conf(VIDEO, "meta.template"))
}
func _baidu_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, BAIDU, name, text, arg...)
	m.Cmdy(mdb.RENDER, web.RENDER.Frame, kit.Format("https://baidu.com/s?wd=%s", text))
}
func _other_show(m *ice.Message, name, text string, arg ...string) {
	_option(m, OTHER, name, text, arg...)
	m.Cmdy(mdb.RENDER, web.RENDER.Frame, text)
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
	IMAGE = "image"
	VIDEO = "video"

	BAIDU = "baidu"
	OTHER = "other"

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
			SPARK: {Name: SPARK, Help: "段落", Value: kit.Data("template", spark)},

			CHART: {Name: CHART, Help: "图表", Value: kit.Data("template", chart, "suffix", `</svg>`)},
			FIELD: {Name: FIELD, Help: "插件", Value: kit.Data("template", field)},
			SHELL: {Name: SHELL, Help: "命令", Value: kit.Data("template", shell)},
			LOCAL: {Name: LOCAL, Help: "文件", Value: kit.Data("template", local)},

			ORDER: {Name: ORDER, Help: "列表", Value: kit.Data("template", order)},
			TABLE: {Name: TABLE, Help: "表格", Value: kit.Data("template", table)},
			IMAGE: {Name: IMAGE, Help: "图片", Value: kit.Data("template", image)},
			VIDEO: {Name: VIDEO, Help: "视频", Value: kit.Data("template", video)},

			WORD: {Name: WORD, Help: "语言文字", Value: kit.Data(
				"path", "usr", "regs", ".*\\.shy", "alias", map[string]interface{}{
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
					ns := strings.Split(cli.NodeName, "-")
					arg = append(arg, kit.Select(ns[len(ns)-1], ""))
				}
				if len(arg) == 1 {
					arg = append(arg, arg[0])
				}
				_title_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			BRIEF: {Name: "brief [name] text", Help: "摘要", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Echo(`<br class="story" data-type="brief">`)
					return
				}
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_brief_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			REFER: {Name: "refer name `[name url]...`", Help: "参考", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_refer_show(m, arg[0], arg[1], arg[2:]...)
			}},
			SPARK: {Name: "spark [name] text", Help: "感悟", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Echo(`<br class="story" data-type="spark">`)
					return
				}
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_spark_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},

			CHART: {Name: "chart label|chain name text arg...", Help: "图表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 2 {
					arg = []string{arg[0], "", arg[1]}
				}
				_chart_show(m, arg[0], arg[1], arg[2], arg[3:]...)
			}},
			FIELD: {Name: "field name cmd", Help: "插件", Action: map[string]*ice.Action{
				"run": {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					if m.Right(arg[1:]) {
						m.Cmdy(arg[1:])
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_field_show(m, strings.ReplaceAll(arg[0], " ", "_"), arg[1], arg[2:]...)
			}},
			SHELL: {Name: "shell [name] cmd", Help: "命令", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_shell_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			LOCAL: {Name: "local [name] text", Help: "文件", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_local_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},

			ORDER: {Name: "order name `[item \n]...`", Help: "列表", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_order_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			TABLE: {Name: "table name `[item item\n]...`", Help: "表格", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if arg[0] == "cmd" {
					msg := m.Cmd(kit.Split(arg[1])).Table()
					arg[1] = msg.Result()
				}

				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_table_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			IMAGE: {Name: "image name url", Help: "图片", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if arg[0] == "qrcode" {
					buf := bytes.NewBuffer(make([]byte, 0, 4096))
					if qr, e := qrcode.New(arg[1], qrcode.Medium); m.Assert(e) {
						m.Assert(qr.Write(kit.Int(kit.Select("256")), buf))
					}
					arg[1] = "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
				}
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_image_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			VIDEO: {Name: "video name url", Help: "视频", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_video_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},

			BAIDU: {Name: "baidu word", Help: "百度", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_baidu_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},
			OTHER: {Name: "other word", Help: "网页", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 1 {
					arg = []string{"", arg[0]}
				}
				_other_show(m, arg[0], kit.Select(arg[0], arg[1]), arg[2:]...)
			}},

			WORD: {Name: "word path=demo/hi.shy auto", Help: "语言文字", Meta: kit.Dict(
				"display", "/plugin/local/wiki/word.js",
			), Action: map[string]*ice.Action{
				web.STORY: {Name: "story", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg[0], "action", "run", arg[1:])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if m.Option(nfs.DIR_DEEP, "true"); reply(m, cmd, arg...) {
					return
				}
				_word_show(m, arg[0])
			}},
		},
	}, nil)
}
