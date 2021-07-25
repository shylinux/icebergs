package wiki

import (
	"fmt"
	"strings"

	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

func _title_show(m *ice.Message, kind, text string, arg ...string) {
	title, _ := m.Optionv(TITLE).(map[string]int)
	switch kind {
	case PREMENU: // 前置目录
		m.RenderTemplate(premenu)
		return

	case ENDMENU: // 后置目录
		m.RenderTemplate(endmenu)
		return

	case SECTION: // 分节标题
		title[SECTION]++
		m.Option("level", "h3")
		m.Option("prefix", fmt.Sprintf("%d.%d ", title[CHAPTER], title[SECTION]))

	case CHAPTER: // 章节标题
		title[CHAPTER]++
		title[SECTION] = 0
		m.Option("level", "h2")
		m.Option("prefix", fmt.Sprintf("%d ", title[CHAPTER]))

	default: // 文章标题
		m.Option("level", "h1")
		m.Option("prefix", "")
	}

	// 添加目录
	menu, _ := m.Optionv("menu").(map[string]interface{})
	menu["list"] = append(menu["list"].([]interface{}), map[string]interface{}{
		"level": m.Option("level"), "prefix": m.Option("prefix"), "content": m.Option("content", text),
	})

	// 渲染引擎
	_wiki_template(m, TITLE, "", text, arg...)
}

const (
	PREMENU = "premenu"
	CHAPTER = "chapter"
	SECTION = "section"
	ENDMENU = "endmenu"
)

const TITLE = "title"

func init() {
	Index.Merge(&ice.Context{
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
		},
		Configs: map[string]*ice.Config{
			TITLE: {Name: TITLE, Help: "标题", Value: kit.Data(
				kit.MDB_TEMPLATE, `<{{.Option "level"}} {{.OptionTemplate}}>{{.Option "prefix"}} {{.Option "content"}}</{{.Option "level"}}>`,
			)},
		},
	})
}

var premenu = `<ul class="story" data-type="premenu"></ul>`
var endmenu = `<ul class="story" data-type="endmenu">{{$menu := .Optionv "menu"}}{{range $index, $value := Value $menu "list"}}
<li>{{Value $value "prefix"}} {{Value $value "content"}}</li>{{end}}
</ul>`
