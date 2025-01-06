package ctx

import (
	"path"
	"reflect"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func isLocalFile(p string) bool {
	return !strings.HasPrefix(p, nfs.PS) && !strings.HasPrefix(p, ice.HTTP)
}
func DisplayTable(m *ice.Message, arg ...ice.Any) *ice.Message {
	return DisplayBase(m, ice.PLUGIN_TABLE_JS, arg...)
}
func DisplayTableCard(m *ice.Message, arg ...ice.Any) *ice.Message {
	return DisplayTable(m, STYLE, "card")
}
func DisplayStory(m *ice.Message, file string, arg ...ice.Any) *ice.Message {
	kit.If(file == "", func() { file = kit.Keys(m.CommandKey(), nfs.JS) })
	kit.If(isLocalFile(file), func() { file = path.Join(ice.PLUGIN_STORY, file) })
	return DisplayBase(m, file, arg...)
}
func DisplayInput(m *ice.Message, file string, arg ...ice.Any) *ice.Message {
	kit.If(file == "", func() { file = kit.Keys(m.CommandKey(), nfs.JS) })
	kit.If(isLocalFile(file), func() { file = path.Join(ice.PLUGIN_INPUT, file) })
	return DisplayBase(m, file, arg...)
}
func DisplayStoryForm(m *ice.Message, arg ...ice.Any) *ice.Message {
	args := kit.List()
	for i := range arg {
		switch v := arg[i].(type) {
		case string:
			args = append(args, ice.SplitCmd("list "+v, nil)...)
		default:
			trans := kit.Value(m.Commands(m.CommandKey()).Meta, ice.CTX_TRANS)
			if t := reflect.TypeOf(v); t.Kind() == reflect.Func {
				name := kit.FuncName(arg[i])
				args = append(args, kit.Dict(mdb.TYPE, html.BUTTON, mdb.NAME, name, mdb.VALUE, kit.Select(name, kit.Value(trans, name), !m.IsEnglish())))
			}
		}
	}
	kit.For(args, func(v ice.Map) { m.Push("", v, kit.Split("type,name,value,values,need,action")) })
	return DisplayStory(m, "form")
}
func DisplayInputKey(m *ice.Message, arg ...ice.Any) *ice.Message {
	return DisplayInput(m, "key", arg...)
}
func DisplayStoryPie(m *ice.Message, arg ...ice.Any) *ice.Message {
	return DisplayStory(m, "pie", arg...)
}
func DisplayStoryJSON(m *ice.Message, arg ...ice.Any) *ice.Message {
	return DisplayStory(m, "json", arg...)
}
func DisplayStorySpide(m *ice.Message, arg ...ice.Any) *ice.Message {
	return DisplayStory(m, "spide", arg...)
}
func DisplayStoryChina(m *ice.Message, arg ...ice.Any) *ice.Message {
	return DisplayStory(m, "china", arg...)
}
func DisplayStudio(m *ice.Message, cmd ...string) *ice.Message {
	for i, k := range cmd {
		kit.If(!strings.Contains(cmd[i], nfs.PT), func() { cmd[i] = m.Prefix(k) })
	}
	return DisplayStory(m.Cmdy(COMMAND, cmd), "studio")
}
func DisplayLocal(m *ice.Message, file string, arg ...ice.Any) *ice.Message {
	kit.If(file == "", func() { file = strings.ReplaceAll(strings.TrimPrefix(m.PrefixKey(), "web."), nfs.PT, nfs.PS) })
	kit.If(isLocalFile(file), func() { file = path.Join(ice.PLUGIN_LOCAL, file) })
	return DisplayBase(m, file, arg...)
}
func DisplayLocalInner(m *ice.Message, arg ...ice.Any) *ice.Message {
	return DisplayLocal(m, "code/inner", arg...)
}
func DisplayBase(m *ice.Message, file string, arg ...ice.Any) *ice.Message {
	m.Option(ice.MSG_DISPLAY, kit.MergeURL(kit.Select(kit.ExtChange(file, nfs.JS), file, strings.Contains(file, mdb.QS)), arg...))
	return Toolkit(m, "")
}
func DisplayBaseCSS(m *ice.Message, file string, arg ...ice.Any) *ice.Message {
	m.Option(ice.MSG_DISPLAY_CSS, kit.MergeURL(kit.Select(kit.ExtChange(file, nfs.CSS), file, strings.Contains(file, mdb.QS)), arg...))
	return m
}
func Toolkit(m *ice.Message, arg ...string) *ice.Message {
	m.OptionDefault(ice.MSG_ONLINE, mdb.Config(m, "online"))
	return m.Options(ice.MSG_TOOLKIT, kit.Select(mdb.Config(m, mdb.TOOLS), kit.Fields(arg)))
}
