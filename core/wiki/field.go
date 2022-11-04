package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func Parse(m *ice.Message, meta string, key string, arg ...string) (data ice.Any) {
	list := []string{}
	for _, line := range kit.SplitLine(strings.Join(arg, ice.SP)) {
		ls := kit.Split(line)
		for i := 0; i < len(ls); i++ {
			if strings.HasPrefix(ls[i], "# ") {
				ls = ls[:i]
				break
			}
		}
		list = append(list, ls...)
	}
	switch data = kit.Parse(nil, "", list...); meta {
	case ice.MSG_OPTION:
		m.Option(key, data)
	case ice.MSG_APPEND:
		m.Append(key, data)
	}
	return data
}
func _field_show(m *ice.Message, name, text string, arg ...string) {
	meta, cmds := kit.Dict(), kit.Split(text)
	m.Search(cmds[0], func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if meta[FEATURE], meta[INPUTS] = kit.Dict(cmd.Meta), cmd.List; name == "" {
			name = cmd.Help
		}
	})
	if len(meta) == 0 || !aaa.Right(m.Spawn(), cmds[0]) {
		return
	}

	name = strings.ReplaceAll(name, ice.SP, "_")
	meta[mdb.NAME], meta[mdb.INDEX] = name, text
	msg := m.Spawn()

	for i := 0; i < len(arg)-1; i += 2 {
		if strings.HasPrefix(arg[i], "opts.") {
			kit.Value(meta, arg[i], m.Option(arg[i], strings.TrimSpace(arg[i+1])))
		} else if strings.HasPrefix(arg[i], "args.") {
			kit.Value(meta, arg[i], m.Option(arg[i], strings.TrimSpace(arg[i+1])))
		} else if strings.HasPrefix(arg[i], ARGS) {
			kit.Value(meta, arg[i], m.Optionv(arg[i], kit.Split(strings.TrimSuffix(strings.TrimPrefix(arg[i+1], "["), "]"))))
		} else {
			kit.Value(meta, arg[i], Parse(m, ice.MSG_OPTION, arg[i], arg[i+1]))
		}

		switch arg[i] {
		case ARGS:
			count := 0
			kit.Fetch(meta[INPUTS], func(index int, value ice.Map) {
				if value[mdb.TYPE] != "button" {
					count++
				}
			})
			if args := kit.Simple(m.Optionv(arg[i])); len(args) > count {
				list := meta[INPUTS].([]ice.Any)
				for i := count; i < len(args); i++ {
					list = append(list, kit.Dict(mdb.TYPE, "text", mdb.NAME, ARGS, mdb.VALUE, args[i]))
				}
				meta[INPUTS] = list
			}
		case TABLE:
			ls := kit.SplitLine(arg[i+1])
			head := kit.SplitWord(ls[0])
			for _, l := range ls[1:] {
				for i, v := range kit.SplitWord(l) {
					msg.Push(head[i], v)
				}
			}
			kit.Value(meta, "msg", msg.FormatMeta(), kit.Keys(FEATURE, "mode"), "simple")
		case SPARK:
			if arg[i+1][0] == '@' && nfs.ExistsFile(m, arg[i+1][1:]) {
				msg.Cmdy(nfs.CAT, arg[i+1][1:])
			} else {
				msg.Echo(strings.TrimSpace(arg[i+1]))
			}
			kit.Value(meta, "msg", msg.FormatMeta(), kit.Keys(FEATURE, "mode"), "simple")

			if text == "web.code.inner" {
				kit.Value(meta, kit.Keys(FEATURE, "display"), "/plugin/local/code/inner.js")
				kit.Value(meta, "plug", kit.UnMarshal(m.Cmdx(mdb.PLUGIN, kit.Ext(name))))
				kit.Value(meta, ARGS, kit.List(path.Dir(name)+ice.PS, path.Base(name)))
			}
		default:
			kit.Value(meta, kit.Keys(FEATURE, arg[i]), msg.Optionv(arg[i], arg[i+1]))
		}
	}
	m.Option(mdb.META, meta)
	_wiki_template(m, name, text)
}

const (
	FEATURE = "feature"
	INPUTS  = "inputs"
	ARGS    = "args"
)
const FIELD = "field"

func init() {
	Index.MergeCommands(ice.Commands{
		FIELD: {Name: "field [name] cmd", Help: "插件", Actions: ice.MergeActions(WordAction(`<fieldset {{.OptionTemplate}}" data-meta='{{.Optionv "meta"|Format}}'>
<legend>{{.Option "name"}}</legend>
<form class="option"></form>
<div class="action"></div>
<div class="output"></div>
<div class="status"></div>
</fieldset>`), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			if arg[1] == ctx.ARGS {
				arg = kit.Simple("", arg)
			}
			if arg = _name(m, arg); strings.Contains(arg[1], ice.NL) {
				arg = kit.Simple(arg[0], "web.chat.div", "auto.cmd", "split", "opts.text", arg[1], arg[2:])
			}
			_field_show(m, arg[0], arg[1], arg[2:]...)
		}},
	})
}
