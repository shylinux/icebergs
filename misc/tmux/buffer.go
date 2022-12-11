package tmux

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	BUFFER = "buffer"
	TEXT   = "text"
)
const (
	SET_BUFFER    = "set-buffer"
	SHOW_BUFFER   = "show-buffer"
	LIST_BUFFER   = "list-buffers"
	DELETE_BUFFER = "delete-buffer"
)

func init() {
	Index.MergeCommands(ice.Commands{
		BUFFER: {Name: "buffer name value auto", Help: "缓存", Actions: ice.Actions{
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {}},
			mdb.CREATE: {Name: "create value*", Hand: func(m *ice.Message, arg ...string) { _tmux_cmd(m, SET_BUFFER, m.Option(mdb.VALUE)) }},
			mdb.REMOVE: {Hand: func(m *ice.Message, arg ...string) { _tmux_cmd(m, DELETE_BUFFER, "-b", m.Option(mdb.NAME)) }},
			mdb.MODIFY: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(arg[0] == mdb.VALUE, func() { _tmux_cmd(m, SET_BUFFER, "-b", m.Option(mdb.NAME), arg[1]) })
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if kit.If(len(arg) > 1 && arg[0] != "" && arg[1] != "", func() { _tmux_cmd(m, SET_BUFFER, "-b", arg[0], arg[1]) }); len(arg) > 0 && arg[0] != "" {
				cli.PushText(m, _tmux_cmds(m, SHOW_BUFFER, "-b", arg[0]))
				return
			}
			for i, v := range kit.SplitLine(_tmux_cmd(m, LIST_BUFFER).Result()) {
				ls := strings.SplitN(v, ": ", 3)
				if m.Push(mdb.NAME, ls[0]).Push(nfs.SIZE, ls[1]); i < 3 {
					m.Push(mdb.VALUE, _tmux_cmd(m, SHOW_BUFFER, "-b", ls[0]).Result())
				} else {
					m.Push(mdb.VALUE, ls[2][1:len(ls[2])-1])
				}
			}
			m.StatusTimeCount().PushAction(mdb.REMOVE).Action(mdb.CREATE)
		}},
		TEXT: {Name: "text auto text:textarea", Help: "文本", Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) > 0, func() { _tmux_cmd(m, SET_BUFFER, arg[0]) })
			cli.PushText(m, _tmux_cmds(m, SHOW_BUFFER))
		}},
	})
}
