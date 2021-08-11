package wiki

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _title_show(m *ice.Message, kind, text string, arg ...string) {
	switch title, _ := m.Optionv(TITLE).(map[string]int); kind {
	case PREMENU: // 前置目录
		_option(m, kind, "", "", arg...)
		m.RenderTemplate(m.Conf(TITLE, kit.Keym(kind)))
		return

	case ENDMENU: // 后置目录
		_option(m, kind, "", "", arg...)
		m.RenderTemplate(m.Conf(TITLE, kit.Keym(kind)))
		return

	case SECTION: // 分节标题
		title[SECTION]++
		m.Option(kit.MDB_LEVEL, "h3")
		m.Option(kit.MDB_PREFIX, kit.Format("%d.%d ", title[CHAPTER], title[SECTION]))

	case CHAPTER: // 章节标题
		title[CHAPTER]++
		title[SECTION] = 0
		m.Option(kit.MDB_LEVEL, "h2")
		m.Option(kit.MDB_PREFIX, kit.Format("%d ", title[CHAPTER]))

	default: // 文章标题
		m.Option(kit.MDB_LEVEL, "h1")
		m.Option(kit.MDB_PREFIX, "")
	}

	// 渲染引擎
	_wiki_template(m, TITLE, "", text, arg...)

	// 添加目录
	menu, _ := m.Optionv(kit.MDB_MENU).(map[string]interface{})
	menu[kit.MDB_LIST] = append(menu[kit.MDB_LIST].([]interface{}), kit.Dict(m.OptionSimple("level,prefix,text")))
}

const (
	PREMENU = "premenu"
	CHAPTER = "chapter"
	SECTION = "section"
	ENDMENU = "endmenu"
)

const TITLE = "title"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TITLE: {Name: "title [premenu|chapter|section|endmenu] text", Help: "标题", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				ns := kit.Split(ice.Info.NodeName, "-")
				arg = append(arg, ns[len(ns)-1])
			}
			switch arg[0] {
			case PREMENU, ENDMENU:
				_title_show(m, arg[0], "", arg[1:]...)
			case CHAPTER, SECTION:
				_title_show(m, arg[0], arg[1], arg[2:]...)
			default:
				_title_show(m, "", arg[0], arg[1:]...)
			}
		}},
	}, Configs: map[string]*ice.Config{
		TITLE: {Name: TITLE, Help: "标题", Value: kit.Data(
			kit.MDB_TEMPLATE, `<{{.Option "level"}} {{.OptionTemplate}}>{{.Option "prefix"}} {{.Option "text"}}</{{.Option "level"}}>`,
			PREMENU, `<ul {{.OptionTemplate}}></ul>`,
			ENDMENU, `<ul {{.OptionTemplate}}>{{$menu := .Optionv "menu"}}
{{range $index, $value := Value $menu "list"}}<li>{{Value $value "prefix"}} {{Value $value "text"}}</li>{{end}}
</ul>`,
		)},
	}})
}
