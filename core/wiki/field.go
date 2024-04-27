package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _field_show(m *ice.Message, name, text string, arg ...string) {
	meta := kit.Dict()
	kit.For(arg, func(k, v string) {
		if k == ctx.ARGS {
			kit.Value(meta, k, kit.Split(strings.TrimSuffix(strings.TrimPrefix(v, "["), "]")))
		} else if k == ice.MSG_RESULT {
			m.Option(html.OUTPUT, strings.TrimSpace(v))
			kit.Value(meta, "meta.mode", ice.MSG_RESULT)
			kit.Value(meta, "msg", kit.Dict())
		} else {
			kit.Value(meta, k, v)
		}
	})
	if p := kit.Format(meta[web.SPACE]); p == "" {
		m.Search(text, func(key string, cmd *ice.Command) {
			meta[mdb.LIST], name = cmd.List, kit.Select(cmd.Help, name)
			kit.For(cmd.Meta, func(k string, v ice.Any) { meta[kit.Keys(mdb.META, k)] = v })
		})
		if m.WarnNotFound(len(meta) == 0, text) || !aaa.Right(m.Spawn(), text) {
			return
		}
	}
	meta[mdb.NAME], meta[mdb.INDEX] = name, text
	_wiki_template(m.Options(mdb.META, kit.Format(meta)), "", kit.Select(name, text, m.IsEnglish()), text)
}

const FIELD = "field"

func init() {
	Index.MergeCommands(ice.Commands{
		FIELD: {Name: "field name cmd", Help: "插件", Hand: func(m *ice.Message, arg ...string) {
			kit.If(kit.IsIn(kit.Select("", arg, 1), ctx.ARGS, web.SPACE), func() { arg = kit.Simple("", arg) })
			arg = _name(m, arg)
			_field_show(m, arg[0], arg[1], arg[2:]...)
		}},
	})
}
