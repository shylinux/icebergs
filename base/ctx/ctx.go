package ctx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
)

func selectAction(list map[string]*ice.Action, fields ...string) map[string]*ice.Action {
	if len(fields) == 0 {
		return list
	}

	res := map[string]*ice.Action{}
	for _, field := range fields {
		res[field] = list[field]
	}
	return res
}
func CmdAction(fields ...string) map[string]*ice.Action {
	return selectAction(map[string]*ice.Action{
		"command": {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
			if !m.PodCmd("command", arg) {
				m.Cmdy("command", arg)
			}
		}},
		"run": {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
			if !m.PodCmd(arg) {
				m.Cmdy(arg)
			}
		}},
	}, fields...)
}

const CTX = "ctx"

var Index = &ice.Context{Name: CTX, Help: "标准模块", Commands: map[string]*ice.Command{
	ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		m.Cmd(mdb.SEARCH, mdb.CREATE, COMMAND, m.Prefix(COMMAND))
	}},
	ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
	}},
}}

func init() { ice.Index.Register(Index, nil, CONTEXT, COMMAND, CONFIG, MESSAGE) }
