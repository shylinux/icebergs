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

type Message interface {
	Option(key string, arg ...ice.Any) string
}

func Display(m Message, file string, arg ...ice.Any) Message {
	if file == "" {
		file = kit.Keys(kit.FileName(2), nfs.JS)
	}
	if !strings.HasPrefix(file, ice.PS) && !strings.HasPrefix(file, ice.HTTP) {
		file = path.Join(ice.PS, path.Join(path.Dir(FileURI(logs.FileLines(2))), file))
	}
	return DisplayBase(m, file, arg...)
}
func DisplayTable(m Message, arg ...ice.Any) Message {
	return DisplayBase(m, "/plugin/table.js", arg...)
}
func DisplayTableCard(m Message, arg ...ice.Any) Message {
	return DisplayTable(m, "style", "card")
}
func DisplayStory(m Message, file string, arg ...ice.Any) Message {
	if file == "" {
		file = kit.ExtChange(kit.FileName(2), nfs.JS)
	}
	if !strings.HasPrefix(file, ice.PS) && !strings.HasPrefix(file, ice.HTTP) {
		file = path.Join(ice.PLUGIN_STORY, file)
	}
	return DisplayBase(m, file, arg...)
}
func DisplayStoryJSON(m Message, arg ...ice.Any) Message {
	return DisplayStory(m, "json", arg...)
}
func DisplayStorySpide(m Message, arg ...ice.Any) Message {
	return DisplayStory(m, "spide", arg...)
}
func DisplayLocal(m Message, file string, arg ...ice.Any) Message {
	if file == "" {
		file = path.Join(kit.PathName(2), kit.Keys(kit.FileName(2), ice.JS))
	}
	if !strings.HasPrefix(file, ice.PS) && !strings.HasPrefix(file, ice.HTTP) {
		file = path.Join(ice.PLUGIN_LOCAL, file)
	}
	return DisplayBase(m, file, arg...)
}
func DisplayBase(m Message, file string, arg ...ice.Any) Message {
	m.Option(ice.MSG_DISPLAY, kit.MergeURL(kit.Select(kit.ExtChange(file, nfs.JS), file, strings.Contains(file, "?")), arg...))
	return m
}
func Toolkit(m *ice.Message, arg ...string) {
	m.Option(ice.MSG_TOOLKIT, kit.Select(mdb.Config(m, mdb.TOOLS), kit.Fields(arg)))
}
