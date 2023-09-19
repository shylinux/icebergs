package ctx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	kit "shylinux.com/x/toolkits"
)

const PROCESS = "process"

var _process = map[string]ice.Any{}

func AddProcess(key string, val ice.Any) { _process[key] = val }
func ProcessAction() ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { AddProcess(m.CommandKey(), m.PrefixKey()) }},
		PROCESS:      {Hand: func(m *ice.Message, arg ...string) { ProcessField(m, m.PrefixKey(), arg, arg...) }},
	}
}

func _process_args(m *ice.Message, args ice.Any) []string {
	switch cb := args.(type) {
	case func() string:
		return []string{cb()}
	case func() []string:
		return cb()
	case []string:
		return cb
	case string:
		return []string{cb}
	case nil:
	default:
		m.ErrorNotImplement(args)
	}
	return nil
}
func Process(m *ice.Message, key string, args ice.Any, arg ...string) {
	switch cb := _process[kit.Select(m.ActionKey(), key)].(type) {
	case string:
		if !kit.HasPrefixList(arg, ACTION, PROCESS) {
			m.Cmdy(cb, PROCESS, _process_args(m, args)).Optionv(ice.FIELD_PREFIX, kit.Simple(m.ActionKey(), m.Optionv(ice.FIELD_PREFIX)))
		} else {
			m.Cmdy(cb, arg)
		}
	default:
		ProcessField(m, key, args, arg...)
	}
}
func ProcessField(m *ice.Message, cmd string, args ice.Any, arg ...string) *ice.Message {
	if cmd = kit.Select(m.ActionKey(), cmd); !kit.HasPrefixList(arg, RUN) {
		if PodCmd(m, COMMAND, cmd) {
			m.Push(ice.SPACE, m.Option(ice.MSG_USERPOD))
		} else {
			m.Cmdy(COMMAND, cmd)
		}
		m.Push(ARGS, kit.Format(_process_args(m, args))).Options(ice.MSG_INDEX, m.PrefixKey()).ProcessField(ACTION, m.ActionKey(), RUN)
	} else {
		if !PodCmd(m, cmd, arg[1:]) {
			kit.If(aaa.Right(m, cmd, arg[1:]), func() { m.Cmdy(cmd, arg[1:]) })
		}
	}
	return m
}
func ProcessCommand(m *ice.Message, cmd string, args []string, arg ...string) {
	if !kit.HasPrefixList(arg, RUN) {
		m.Cmdy(COMMAND, cmd).Push(ice.ARG, kit.Format(args)).ProcessField(cmd, RUN)
	} else {
		m.Cmdy(cmd, arg[1:])
	}
}
