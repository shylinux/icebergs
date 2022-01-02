package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func Parse(m *ice.Message, meta string, key string, arg ...string) *ice.Message {
	list := []string{}
	for _, line := range kit.Split(strings.Join(arg, ice.SP), ice.NL) {
		ls := kit.Split(line)
		for i := 0; i < len(ls); i++ {
			if strings.HasPrefix(ls[i], "#") {
				ls = ls[:i]
				break
			}
		}
		list = append(list, ls...)
	}

	switch data := kit.Parse(nil, "", list...); meta {
	case ice.MSG_OPTION:
		m.Option(key, data)
	case ice.MSG_APPEND:
		m.Append(key, data)
	}
	return m
}
func _field_show(m *ice.Message, name, text string, arg ...string) {
	// 命令参数
	meta, cmds := kit.Dict(), kit.Split(text)
	m.Search(cmds[0], func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if meta[FEATURE], meta[INPUTS] = cmd.Meta, cmd.List; name == "" {
			name = cmd.Help
		}
	})

	name = strings.ReplaceAll(name, ice.SP, "_")
	meta[mdb.NAME] = name
	meta[mdb.INDEX] = text

	// 扩展参数
	for i := 0; i < len(arg)-1; i += 2 {
		if strings.HasPrefix(arg[i], "opts.") {
			m.Option(arg[i], strings.TrimSpace(arg[i+1]))
			kit.Value(meta, arg[i], m.Option(arg[i]))
		} else if strings.HasPrefix(arg[i], "args.") {
			m.Option(arg[i], strings.TrimSpace(arg[i+1]))
			kit.Value(meta, arg[i], m.Option(arg[i]))
		} else if strings.HasPrefix(arg[i], ARGS) {
			m.Option(arg[i], kit.Split(strings.TrimSuffix(strings.TrimPrefix(arg[i+1], "["), "]")))
			kit.Value(meta, arg[i], m.Optionv(arg[i]))
		} else {
			Parse(m, ice.MSG_OPTION, arg[i], arg[i+1])
			kit.Value(meta, arg[i], m.Optionv(arg[i]))
		}

		switch arg[i] {
		case kit.MDB_CONTENT:
			meta[arg[i]] = arg[i+1]

		case ARGS:
			args := kit.Simple(m.Optionv(arg[i]))

			count := 0
			kit.Fetch(meta[INPUTS], func(index int, value map[string]interface{}) {
				if value[mdb.TYPE] != kit.MDB_BUTTON {
					count++
				}
			})

			if len(args) > count {
				list := meta[INPUTS].([]interface{})
				for i := count; i < len(args); i++ {
					list = append(list, kit.Dict(
						mdb.TYPE, "text", mdb.NAME, "args", mdb.VALUE, args[i],
					))
				}
				meta[INPUTS] = list
			}
		}
	}
	m.Option(mdb.META, meta)

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
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		FIELD: {Name: "field [name] cmd", Help: "插件", Action: ice.MergeAction(map[string]*ice.Action{
			ice.RUN: {Name: "run", Help: "执行"},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if arg = _name(m, arg); strings.Contains(arg[1], ice.NL) {
				arg = append([]string{arg[0], "web.chat.div", "auto.cmd", "split", "opts.text", arg[1]}, arg[2:]...)
			}
			_field_show(m, arg[0], arg[1], arg[2:]...)
		}},
	}, Configs: map[string]*ice.Config{
		FIELD: {Name: FIELD, Help: "插件", Value: kit.Data(
			nfs.TEMPLATE, `<fieldset {{.OptionTemplate}}" data-meta='{{.Optionv "meta"|Format}}'>
<legend>{{.Option "name"}}</legend>
<form class="option"></form>
<div class="action"></div>
<div class="output"></div>
<div class="status"></div>
</fieldset>`,
		)},
	}})
}
