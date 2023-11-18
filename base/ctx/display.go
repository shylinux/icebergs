package ctx

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/logs"
)

type displayMessage interface {
	Option(key string, arg ...ice.Any) string
	Action(arg ...ice.Any) *ice.Message
}

func isLocalFile(p string) bool {
	return !strings.HasPrefix(p, nfs.PS) && !strings.HasPrefix(p, ice.HTTP)
}
func Display(m *ice.Message, file string, arg ...ice.Any) *ice.Message {
	kit.If(file == "", func() { file = kit.Keys(kit.FileName(5), nfs.JS) })
	kit.If(isLocalFile(file), func() { file = path.Join(nfs.PS, path.Join(path.Dir(FileURI(logs.FileLines(2))), file)) })
	return DisplayBase(m, file, arg...)
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
	return DisplayStory(m.Cmdy(COMMAND, cmd), "studio.js")
}
func DisplayLocal(m *ice.Message, file string, arg ...ice.Any) *ice.Message {
	kit.If(file == "", func() { file = path.Join(kit.PathName(5), kit.Keys(kit.FileName(5), nfs.JS)) })
	kit.If(isLocalFile(file), func() { file = path.Join(ice.PLUGIN_LOCAL, file) })
	return DisplayBase(m, file, arg...)
}
func DisplayLocalInner(m *ice.Message, arg ...ice.Any) *ice.Message {
	return DisplayLocal(m, "code/inner.js", arg...)
}
func DisplayBase(m *ice.Message, file string, arg ...ice.Any) *ice.Message {
	m.Option(ice.MSG_DISPLAY, kit.MergeURL(kit.Select(kit.ExtChange(file, nfs.JS), file, strings.Contains(file, mdb.QS)), arg...))
	Toolkit(m, "")
	return m
}
func Toolkit(m *ice.Message, arg ...string) {
	m.Option(ice.MSG_TOOLKIT, kit.Select(mdb.Config(m, mdb.TOOLS), kit.Fields(arg)))
}
