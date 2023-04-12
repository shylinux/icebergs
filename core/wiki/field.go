package wiki

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func _field_show(m *ice.Message, name, text string, arg ...string) {
	meta := kit.Dict()
	m.Search(text, func(key string, cmd *ice.Command) {
		meta[mdb.META], meta[mdb.LIST], name = kit.Dict(cmd.Meta), cmd.List, kit.Select(cmd.Help, name)
	})
	if m.Warn(len(meta) == 0, ice.ErrNotFound, text) || !aaa.Right(m.Spawn(), text) {
		return
	}
	kit.For(arg, func(k, v string) {
		if k == ctx.ARGS {
			kit.Value(meta, k, kit.Split(strings.TrimSuffix(strings.TrimPrefix(v, "["), "]")))
		} else {
			kit.Value(meta, k, v)
		}
	})
	meta[mdb.NAME], meta[mdb.INDEX] = name, text
	_wiki_template(m.Options(mdb.META, kit.Format(meta)), "", name, text)
}

const FIELD = "field"

func init() {
	Index.MergeCommands(ice.Commands{
		FIELD: {Name: "field name cmd", Help: "插件", Actions: ctx.CmdAction(), Hand: func(m *ice.Message, arg ...string) {
			kit.If(kit.Select("", arg, 1) == ctx.ARGS, func() { arg = kit.Simple("", arg) })
			arg = _name(m, arg)
			_field_show(m, arg[0], arg[1], arg[2:]...)
		}},
	})
}
