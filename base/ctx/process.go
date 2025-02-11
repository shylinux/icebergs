package ctx

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _process_args(m *ice.Message, args ice.Any) []string {
	switch cb := args.(type) {
	case []string:
		return cb
	case string:
		return []string{cb}
	case func() string:
		return []string{cb()}
	case func() []string:
		return cb()
	case func():
		cb()
	case nil:
	default:
		m.ErrorNotImplement(args)
	}
	return nil
}
func ProcessField(m *ice.Message, cmd string, args ice.Any, arg ...string) *ice.Message {
	if cmd = kit.Select(m.ActionKey(), cmd); kit.HasPrefixList(arg, RUN) {
		if !PodCmd(m, cmd, arg[1:]) && aaa.Right(m, cmd, arg[1:]) {
			m.Cmdy(cmd, arg[1:])
		}
		return m
	}
	args = kit.Format(_process_args(m, args))
	if PodCmd(m, COMMAND, cmd) {
		m.Push(ice.SPACE, m.Option(ice.MSG_USERPOD))
	} else {
		m.Cmdy(COMMAND, cmd)
	}
	if m.Push(ARGS, args); m.IsMetaKey() {
		m.Push(STYLE, html.FLOAT)
	}
	if m.ActionKey() == "" {
		m.ProcessField(ACTION, RUN, cmd)
	} else {
		m.ProcessField(ACTION, m.ActionKey(), RUN)
	}
	return m
}
func ProcessFloat(m *ice.Message, cmd string, args ice.Any, arg ...string) *ice.Message {
	if m.IsMetaKey() {
		m.ProcessOpen(path.Join("/c/", cmd, path.Join(_process_args(m, args)...)))
		return m
	} else if !kit.HasPrefixList(arg, RUN) {
		defer m.Push(STYLE, html.FLOAT)
	}
	return ProcessField(m, cmd, args, arg...)
}
