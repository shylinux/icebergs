package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _title_parse(m *ice.Message, dir string, text string) string {
	return m.Cmdx(lex.SPLIT, "", "name,link", kit.Dict(nfs.CAT_CONTENT, text), func(ls []string, data ice.Map) []string {
		if len(ls) > 1 {
			ls[1] = path.Join(dir, ls[1])
		}
		return ls
	})
}

func _title_show(m *ice.Message, kind, text string, arg ...string) {
	switch title, _ := m.Optionv(TITLE).(map[string]int); kind {
	case NAVMENU: // 导航目录
		_option(m, kind, "", text, arg...)
		data := _title_parse(m, path.Dir(m.Option(ice.MSG_SCRIPT)), text)
		m.RenderTemplate(kit.Format("<div {{.OptionTemplate}} data-data='%s'></div>", data))
		return

	case PREMENU: // 前置目录
		_option(m, kind, "", "", arg...)
		m.RenderTemplate(m.Config(kind))
		return

	case ENDMENU: // 后置目录
		_option(m, kind, "", "", arg...)
		m.RenderTemplate(m.Config(kind))
		return

	case SECTION: // 分节标题
		title[SECTION]++
		m.Option(LEVEL, "h3")
		m.Option(PREFIX, kit.Format("%d.%d ", title[CHAPTER], title[SECTION]))

	case CHAPTER: // 章节标题
		title[CHAPTER]++
		title[SECTION] = 0
		m.Option(LEVEL, "h2")
		m.Option(PREFIX, kit.Format("%d ", title[CHAPTER]))

	default: // 文章标题
		m.Option(LEVEL, "h1")
		m.Option(PREFIX, "")
	}

	// 渲染引擎
	_wiki_template(m, TITLE, "", text, arg...)

	// 添加目录
	menu, _ := m.Optionv(MENU).(ice.Map)
	menu[mdb.LIST] = append(menu[mdb.LIST].([]ice.Any), kit.Dict(m.OptionSimple("level,prefix,text")))
}

const (
	NAVMENU = "navmenu"
	PREMENU = "premenu"
	CHAPTER = "chapter"
	SECTION = "section"
	ENDMENU = "endmenu"
)

const (
	PREFIX = "prefix"
	LEVEL  = "level"
	MENU   = "menu"
)
const TITLE = "title"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		TITLE: {Name: "title [navmenu|premenu|chapter|section|endmenu] text", Help: "标题", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				arg = append(arg, kit.Slice(kit.Split(ice.Info.NodeName, "-"), -1)[0])
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
	}, Configs: ice.Configs{
		TITLE: {Name: TITLE, Help: "标题", Value: kit.Data(
			nfs.TEMPLATE, `<{{.Option "level"}} {{.OptionTemplate}}>{{.Option "prefix"}} {{.Option "text"}}</{{.Option "level"}}>`,
			PREMENU, `<ul {{.OptionTemplate}}></ul>`,
			ENDMENU, `<ul {{.OptionTemplate}}>{{$menu := .Optionv "menu"}}
{{range $index, $value := Value $menu "list"}}<li>{{Value $value "prefix"}} {{Value $value "text"}}</li>{{end}}
</ul>`,
		)},
	}})
}
