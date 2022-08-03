package ctx

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func ProcessCommand(m *ice.Message, cmd string, val []string, arg ...string) {
	if len(arg) > 0 && arg[0] == ice.RUN {
		m.Cmdy(cmd, arg[1:])
		return
	}

	m.Cmdy(COMMAND, cmd)
	m.ProcessField(cmd, ice.RUN)
	m.Push(ice.ARG, kit.Format(val))
}
func ProcessCommandOpt(m *ice.Message, arg []string, args ...string) {
	if len(arg) > 0 && arg[0] == ice.RUN {
		return
	}
	m.Push("opt", kit.Format(m.OptionSimple(args...)))
}
func DisplayTable(m *ice.Message, arg ...ice.Any) *ice.Message { // /plugin/story/file
	return m.Display(kit.MergeURL("/plugin/table.js", arg...))
}
func DisplayTableCard(m *ice.Message, arg ...ice.Any) *ice.Message { // /plugin/story/file
	return m.Display(kit.MergeURL("/plugin/table.js", "style", "card"))
}
func DisplayStory(m *ice.Message, file string, arg ...ice.Any) *ice.Message { // /plugin/story/file
	if !strings.HasPrefix(file, ice.HTTP) && !strings.HasPrefix(file, ice.PS) {
		file = path.Join(ice.PLUGIN_STORY, file)
	}
	return DisplayBase(m, file, arg...)
}
func DisplayStoryJSON(m *ice.Message, arg ...ice.Any) *ice.Message { // /plugin/story/json.js
	return DisplayStory(m, "json", arg...)
}
func DisplayStorySpide(m *ice.Message, arg ...ice.Any) *ice.Message { // /plugin/story/json.js
	return DisplayStory(m, "spide", arg...).StatusTimeCount()
}
func DisplayLocal(m *ice.Message, file string, arg ...ice.Any) *ice.Message { // /plugin/local/file
	if file == "" {
		file = path.Join(kit.PathName(2), kit.Keys(kit.FileName(2), ice.JS))
	}
	if !strings.HasPrefix(file, ice.HTTP) && !strings.HasPrefix(file, ice.PS) {
		file = path.Join(ice.PLUGIN_LOCAL, file)
	}
	return DisplayBase(m, file, arg...)
}
func DisplayBase(m *ice.Message, file string, arg ...ice.Any) *ice.Message {
	if !strings.Contains(file, ice.PT) {
		file += ".js"
	}
	m.Option(ice.MSG_DISPLAY, kit.MergeURL(ice.DisplayBase(file)[ice.DISPLAY], arg...))
	return m
}
