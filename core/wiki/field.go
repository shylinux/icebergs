package wiki

import (
	"path"
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	kit "github.com/shylinux/toolkits"
)

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
	m.Option(kit.MDB_META, data)
	m.RenderTemplate(m.Conf(FIELD, kit.Keym(kit.MDB_TEMPLATE)))
}

const FIELD = "field"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			FIELD: {Name: "field [name] cmd", Help: "插件", Action: map[string]*ice.Action{
				cli.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					if !m.Warn(!m.Right(arg[1:]), ice.ErrNotRight, arg[1:]) {
						m.Cmdy(arg[1:])
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_field_show(m, strings.ReplaceAll(kit.Select(path.Base(arg[1]), arg[0]), " ", "_"), arg[1], arg[2:]...)
			}},
		},
		Configs: map[string]*ice.Config{
			FIELD: {Name: FIELD, Help: "插件", Value: kit.Data(
				kit.MDB_TEMPLATE, `<fieldset {{.OptionTemplate}}" data-meta='{{.Optionv "meta"|Format}}'>
<legend>{{.Option "name"}}</legend>
<form class="option"></form>
<div class="action"></div>
<div class="output"></div>
<div class="status"></div>
</fieldset>`,
			)},
		},
	})
}
