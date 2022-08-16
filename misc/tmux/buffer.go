package tmux

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const (
	BUFFER = "buffer"
	TEXT   = "text"
)
const (
	SET_BUFFER  = "set-buffer"
	SHOW_BUFFER = "show-buffer"
	LIST_BUFFER = "list-buffers"
)

func init() {
	Index.MergeCommands(ice.Commands{
		BUFFER: {Name: "buffer name value auto", Help: "缓存", Actions: ice.Actions{
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case mdb.TEXT:
					_tmux_cmd(m, SET_BUFFER, "-b", m.Option(mdb.NAME), arg[1])
				}
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 1 && arg[1] != "" { // 设置缓存
				_tmux_cmd(m, SET_BUFFER, "-b", arg[0], arg[1])
			}
			if len(arg) > 0 { // 查看缓存
				m.Echo(_tmux_cmd(m, SHOW_BUFFER, "-b", arg[0]).Result())
				return
			}

			// 缓存列表
			for i, v := range kit.Split(_tmux_cmd(m, LIST_BUFFER).Result(), ice.NL, ice.NL, ice.NL) {
				ls := strings.SplitN(v, ": ", 3)
				m.Push(mdb.NAME, ls[0])
				m.Push(nfs.SIZE, ls[1])
				if i < 3 {
					m.Push(mdb.TEXT, _tmux_cmd(m, SHOW_BUFFER, "-b", ls[0]).Result())
				} else {
					m.Push(mdb.TEXT, ls[2][1:len(ls[2])-1])
				}
			}
		}},
		TEXT: {Name: "text auto save text:textarea", Help: "文本", Actions: ice.Actions{
			nfs.SAVE: {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
				if len(arg) > 0 && arg[0] != "" {
					_tmux_cmd(m, SET_BUFFER, arg[0])
				}
				m.Cmdy(TEXT)
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			text := _tmux_cmd(m, SHOW_BUFFER).Result()
			if m.EchoQRCode(text).Echo(ice.NL); strings.HasPrefix(text, ice.HTTP) {
				m.EchoAnchor(text)
			} else {
				m.EchoScript(text)
			}
		}},
	})
}
