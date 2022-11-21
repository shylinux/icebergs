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
func ProcessField(m *ice.Message, cmd string, args []string, arg ...string) {
	cmd = kit.Select(m.PrefixKey(), cmd)
	if len(arg) == 0 || arg[0] != ice.RUN {
		m.Option("_index", m.PrefixKey())
		if m.Cmdy(COMMAND, cmd).ProcessField(ACTION, m.ActionKey(), ice.RUN); len(args) > 0 {
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

func ProcessRefresh(m *ice.Message, arg ...string) { m.ProcessRefresh(arg...) }
func ProcessRewrite(m *ice.Message, arg ...ice.Any) { m.ProcessRewrite(arg...) }
func ProcessHold(m *ice.Message, text ...ice.Any) { m.Process(ice.PROCESS_HOLD, text...) }
func ProcessOpen(m *ice.Message, url string) { m.Process(ice.PROCESS_OPEN, url) }

func ProcessAction() ice.Actions {
	return ice.Actions{
		ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { AddProcess(m.CommandKey(), m.PrefixKey()) }},
		PROCESS: {Hand: func(m *ice.Message, arg ...string) { ProcessField(m, m.PrefixKey(), arg, arg...) }},
	}
}
