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
func ProcessField(m *ice.Message, cmd string, args []string, arg ...string) {
	cmd = kit.Select(m.PrefixKey(), cmd)
	if len(arg) == 0 || arg[0] != ice.RUN {
		if m.Cmdy(COMMAND, cmd).ProcessField(m.ActionKey(), ice.RUN); len(args) > 0 {
			m.Push(ARGS, kit.Format(args))
		}
		return
	}
	if aaa.Right(m, cmd, arg[1:]) {
		m.Cmdy(cmd, arg[1:])
	}
}
func ProcessFloat(m *ice.Message, arg ...string) {
	m.Option(ice.MSG_PROCESS, ice.PROCESS_FLOAT)
	m.Option(ice.PROCESS_ARG, arg)
	m.Cmdy(COMMAND, arg[0])
}

func ProcessHold(m *ice.Message, text ...ice.Any) {
	m.Process(ice.PROCESS_HOLD, text...)
}
func ProcessRefresh(m *ice.Message, arg ...string) {
	m.ProcessRefresh(kit.Select("300ms", arg, 0))
}
func ProcessRewrite(m *ice.Message, arg ...ice.Any) {
	m.ProcessRewrite(arg...)
}

func ProcessAction() ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
			AddProcess(m.CommandKey(), m.PrefixKey())
		}},
		PROCESS: {Name: "process", Help: "响应", Hand: func(m *ice.Message, arg ...string) {
			ProcessField(m, m.PrefixKey(), arg, arg...)
		}},
	}
}
