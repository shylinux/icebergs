package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _title_deep(str string) int {
	for i, c := range str {
		if c != ' ' {
			return i
		}
	}
	return 0
}
func _title_parse(m *ice.Message, dir string, root map[string]interface{}, list []string) int {
	var last map[string]interface{}
	deep := _title_deep(list[0])
	for i := 0; i < len(list); i++ {
		if d := _title_deep(list[i]); d < deep {
			return i
		} else if d > deep {
			i += _title_parse(m, dir, last, list[i:]) - 1
			continue
		}

		ls := kit.Split(list[i])
		switch len(ls) {
		case 0:
			continue
		case 1:
		default:
			ls[1] = path.Join(dir, ls[1])
		}

		meta := kit.Dict(
			"name", kit.Select("", ls, 0),
			"link", kit.Select("", ls, 1),
		)
		for i := 2; i < len(ls); i += 2 {
			meta[ls[i]] = ls[i+1]
		}
		last = kit.Dict("meta", meta, "list", kit.List())
		kit.Value(root, "list.-2", last)
	}
	return len(list)
}

func _title_show(m *ice.Message, kind, text string, arg ...string) {
	switch title, _ := m.Optionv(TITLE).(map[string]int); kind {
	case NAVMENU: // 导航目录
		_option(m, kind, "", text, arg...)
		data := kit.Dict("meta", kit.Dict(), "list", kit.List())
		_title_parse(m, path.Dir(m.Option(ice.MSG_SCRIPT)), data, strings.Split(text, ice.NL))
		m.RenderTemplate(kit.Format("<div {{.OptionTemplate}} data-data='%s'></div>", kit.Format(data)))
		return

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
	NAVMENU = "navmenu"
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
			case NAVMENU:
				_title_show(m, arg[0], arg[1], arg[2:]...)
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
