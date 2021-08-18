package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	kit "shylinux.com/x/toolkits"
)

func _field_show(m *ice.Message, name, text string, arg ...string) {
	// 命令参数
	meta, cmds := kit.Dict(), kit.Split(text)
	m.Search(cmds[0], func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if meta[FEATURE], meta[INPUTS] = cmd.Meta, cmd.List; name == "" {
			name = cmd.Help
		}
	})

	name = strings.ReplaceAll(name, " ", "_")
	meta[kit.MDB_NAME] = name
	meta[kit.MDB_INDEX] = text

	// 扩展参数
	for i := 0; i < len(arg)-1; i += 2 {
		if strings.HasPrefix(arg[i], "args.") {
			m.Option(arg[i], strings.TrimSpace(arg[i+1]))
			kit.Value(meta, arg[i], m.Option(arg[i]))
		} else if strings.HasPrefix(arg[i], ARGS) {
			m.Option(arg[i], kit.Split(strings.TrimSuffix(strings.TrimPrefix(arg[i+1], "["), "]")))
			kit.Value(meta, arg[i], m.Optionv(arg[i]))
		} else {
			m.Parse(ice.MSG_OPTION, arg[i], arg[i+1])
			kit.Value(meta, arg[i], m.Optionv(arg[i]))
		}

		switch arg[i] {
		case kit.MDB_CONTENT:
			meta[arg[i]] = arg[i+1]

		case ARGS:
			args := kit.Simple(m.Optionv(arg[i]))

			count := 0
			kit.Fetch(meta[INPUTS], func(index int, value map[string]interface{}) {
				if value[kit.MDB_INPUT] != kit.MDB_BUTTON && value[kit.MDB_TYPE] != kit.MDB_BUTTON {
					count++
				}
			})

			if len(args) > count {
				list := meta[INPUTS].([]interface{})
				for i := count; i < len(args); i++ {
					list = append(list, kit.Dict(
						kit.MDB_INPUT, "text", kit.MDB_NAME, "args", kit.MDB_VALUE, args[i],
					))
				}
				meta[INPUTS] = list
			}
		}
	}
	m.Option(kit.MDB_META, meta)

	// 渲染引擎
	_wiki_template(m, FIELD, name, text)
}

const (
	FEATURE = "feature"
	INPUTS  = "inputs"
	ARGS    = "args"
)
const FIELD = "field"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			FIELD: {Name: "field [name] cmd", Help: "插件", Action: map[string]*ice.Action{
				cli.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				arg = _name(m, arg)
				_field_show(m, arg[0], arg[1], arg[2:]...)
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
