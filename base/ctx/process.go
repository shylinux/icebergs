package ctx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	kit "shylinux.com/x/toolkits"
)

const PROCESS = "process"

var _process = map[string]ice.Any{}

func AddProcess(key string, val ice.Any) { _process[key] = val }
func Process(m *ice.Message, key string, args []string, arg ...string) {
	switch cb := _process[key].(type) {
	case func(*ice.Message, []string, ...string):
		cb(m, args, arg...)
	case string:
		if len(arg) == 0 || arg[0] != PROCESS {
			m.Cmdy(cb, PROCESS, args)
			m.Optionv(ice.FIELD_PREFIX, kit.Simple(m.ActionKey(), m.Optionv(ice.FIELD_PREFIX)))
		} else {
			m.Cmdy(cb, arg)
		}
	default:
		ProcessField(m, key, args, arg...)
	}
}
func ProcessCommand(m *ice.Message, cmd string, val []string, arg ...string) {
	if len(arg) > 0 && arg[0] == ice.RUN {
		m.Cmdy(cmd, arg[1:])
	} else {
		m.Cmdy(COMMAND, cmd).Push(ice.ARG, kit.Format(val))
		m.ProcessField(cmd, ice.RUN)
	}
}
func ProcessCommandOpt(m *ice.Message, arg []string, args ...string) {
	if len(arg) > 0 && arg[0] == ice.RUN {
		return
	}
	m.Push("opt", kit.Format(m.OptionSimple(args...)))
}
func ProcessFloat(m *ice.Message, arg ...string) {
	m.Option(ice.MSG_PROCESS, ice.PROCESS_FLOAT)
	m.Option(ice.PROCESS_ARG, arg)
	m.Cmdy(COMMAND, arg[0])
}
func ProcessField(m *ice.Message, cmd string, args ice.Any, arg ...string) {
	if cmd = kit.Select(m.ActionKey(), cmd); len(arg) == 0 || arg[0] != ice.RUN {
		m.Option("_index", m.PrefixKey())
		m.Cmdy(COMMAND, cmd).ProcessField(ACTION, m.ActionKey(), ice.RUN)
		switch cb := args.(type) {
		case func() string:
			m.Push(ARGS, kit.Format([]string{cb()}))
		case func() []string:
			m.Push(ARGS, kit.Format(cb()))
		case []string:
			m.Push(ARGS, kit.Format(cb))
		case string:
			m.Push(ARGS, kit.Format([]string{cb}))
		case nil:
		default:
			m.ErrorNotImplement(args)
		}
	} else {
		if aaa.Right(m, cmd, arg[1:]) {
			m.Cmdy(cmd, arg[1:])
		}
	}
}

func ProcessRefresh(m *ice.Message, arg ...string)  { m.ProcessRefresh(arg...) }
func ProcessRewrite(m *ice.Message, arg ...ice.Any) { m.ProcessRewrite(arg...) }
func ProcessHold(m *ice.Message, text ...ice.Any)   { m.Process(ice.PROCESS_HOLD, text...) }
func ProcessOpen(m *ice.Message, url string)        { m.Process(ice.PROCESS_OPEN, url) }

func ProcessAction() ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { AddProcess(m.CommandKey(), m.PrefixKey()) }},
		PROCESS:      {Hand: func(m *ice.Message, arg ...string) { ProcessField(m, m.PrefixKey(), arg, arg...) }},
	}
}
