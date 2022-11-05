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
	return m.Cmdx(lex.SPLIT, "", "name,link", kit.Dict(nfs.CAT_CONTENT, text), func(ls []string) []string {
		if len(ls) > 1 {
			ls[1] = path.Join(dir, ls[1])
		}
		return ls
	})
}
func _title_menu(m *ice.Message, kind, text string, arg ...string) *ice.Message {
	if kind == NAVMENU {
		m.Option(mdb.DATA, _title_parse(m, path.Dir(m.Option(ice.MSG_SCRIPT)), text))
	}
	return _option(m, kind, "", text, arg...).RenderTemplate(m.Config(kind), &Message{m})
}
func _title_show(m *ice.Message, kind, text string, arg ...string) *ice.Message {
	switch title, _ := m.Optionv(TITLE).(map[string]int); kind {
	case SECTION:
		title[SECTION]++
		m.Option(LEVEL, "h3")
		m.Option(PREFIX, kit.Format("%d.%d ", title[CHAPTER], title[SECTION]))
	case CHAPTER:
		title[CHAPTER]++
		title[SECTION] = 0
		m.Option(LEVEL, "h2")
		m.Option(PREFIX, kit.Format("%d ", title[CHAPTER]))
	default:
		m.Option(LEVEL, "h1")
		m.Option(PREFIX, "")
	}
	_wiki_template(m, "", text, arg...)
	menu, _ := m.Optionv(MENU).(ice.Map)
	menu[mdb.LIST] = append(menu[mdb.LIST].([]ice.Any), kit.Dict(m.OptionSimple("level,prefix,text")))
	return m
}

const (
	PREFIX = "prefix"
	LEVEL  = "level"
	MENU   = "menu"
)
const (
	NAVMENU = "navmenu"
	PREMENU = "premenu"
	CHAPTER = "chapter"
	SECTION = "section"
	ENDMENU = "endmenu"
)
const TITLE = "title"

func init() {
	Index.MergeCommands(ice.Commands{
		TITLE: {Name: "title [navmenu|premenu|chapter|section|endmenu] text", Actions: WordAction(
			`<{{.Option "level"}} {{.OptionTemplate}}>{{.Option "prefix"}} {{.Option "text"}}</{{.Option "level"}}>`,
			NAVMENU, `<div {{.OptionTemplate}} data-data='{{.Option "data"}}'></div>`,
			PREMENU, `<ul {{.OptionTemplate}}></ul>`,
			ENDMENU, `<ul {{.OptionTemplate}}>{{$menu := .Optionv "menu"}}
{{range $index, $value := Value $menu "list"}}<li>{{Value $value "prefix"}} {{Value $value "text"}}</li>{{end}}
</ul>`), Help: "标题", Hand: func(m *ice.Message, arg ...string) {
			switch arg[0] {
			case NAVMENU:
				_title_menu(m, arg[0], arg[1], arg[2:]...)
			case PREMENU, ENDMENU:
				_title_menu(m, arg[0], "", arg[1:]...)
			case CHAPTER, SECTION:
				_title_show(m, arg[0], arg[1], arg[2:]...)
			default:
				_title_show(m, "", arg[0], arg[1:]...)
			}
		}},
	})
}
