package wiki

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _title_parse(m *ice.Message, text string) string {
	return m.Cmdx(lex.SPLIT, "", "name,link", kit.Dict(nfs.CAT_CONTENT, text), func(ls []string) []string {
		kit.If(!kit.HasPrefix(ls[1], nfs.PS, web.HTTP), func() { ls[1] = path.Join(path.Dir(m.Option(ice.MSG_SCRIPT)), ls[1]) })
		return ls
	})
}
func _title_menu(m *ice.Message, name, text string, arg ...string) *ice.Message {
	m.Options(mdb.DATA, _title_parse(m, text))
	return _wiki_template(m, name, name, text, arg...)
}
func _title_show(m *ice.Message, name, text string, arg ...string) *ice.Message {
	switch title, _ := m.Optionv(TITLE).(map[string]int); name {
	case SECTION:
		title[SECTION]++
		m.Options(LEVEL, "h3", PREFIX, kit.Format("%d.%d ", title[CHAPTER], title[SECTION]))
	case CHAPTER:
		title[CHAPTER]++
		title[SECTION] = 0
		m.Options(LEVEL, "h2", PREFIX, kit.Format("%d ", title[CHAPTER]))
	default:
		m.Options(LEVEL, "h1", PREFIX, "")
	}
	return _wiki_template(m, "", name, text, arg...)
}

const (
	LEVEL  = "level"
	PREFIX = "prefix"
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
		TITLE: {Name: "title type=navmenu,premenu,chapter,section,endmenu text", Help: "标题", Hand: func(m *ice.Message, arg ...string) {
			switch arg[0] {
			case NAVMENU: // navmenu text arg...
				_title_menu(m, arg[0], arg[1], arg[2:]...)
			case PREMENU, ENDMENU: // premenu arg...
				_title_menu(m, arg[0], "", arg[1:]...)
			case CHAPTER, SECTION: // chapter text arg...
				_title_show(m, arg[0], arg[1], arg[2:]...)
			default: // title text arg...
				_title_show(m, "", arg[0], arg[1:]...)
			}
		}},
	})
}
