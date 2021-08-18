package tmux

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			BUFFER: {Name: BUFFER, Help: "缓存", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			BUFFER: {Name: "buffer name value auto export import", Help: "缓存", Action: map[string]*ice.Action{
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_TEXT:
						m.Cmd(cli.SYSTEM, TMUX, "set-buffer", "-b", m.Option(kit.MDB_NAME), arg[1])
					}
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Conf(BUFFER, mdb.LIST, "")
					m.Conf(BUFFER, kit.Keys(mdb.META, kit.MDB_COUNT), "0")

					m.Cmd(BUFFER).Table(func(index int, value map[string]string, head []string) {
						m.Grow(BUFFER, "", kit.Dict(
							kit.MDB_NAME, value[head[0]], kit.MDB_TEXT, m.Cmdx(cli.SYSTEM, TMUX, "show-buffer", "-b", value[head[0]]),
						))
					})
					m.Cmdy(mdb.EXPORT, m.Prefix(BUFFER), "", mdb.LIST)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Conf(BUFFER, mdb.LIST, "")
					m.Conf(BUFFER, kit.Keys(mdb.META, kit.MDB_COUNT), "0")

					m.Cmdy(mdb.IMPORT, m.Prefix(BUFFER), "", mdb.LIST)
					m.Grows(BUFFER, "", "", "", func(index int, value map[string]interface{}) {
						m.Cmd(cli.SYSTEM, TMUX, "set-buffer", "-b", value[kit.MDB_NAME], value[kit.MDB_TEXT])
					})
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 1 && arg[1] != "" { // 设置缓存
					m.Cmd(cli.SYSTEM, TMUX, "set-buffer", "-b", arg[0], arg[1])
				}
				if len(arg) > 0 { // 查看缓存
					m.Echo(m.Cmdx(cli.SYSTEM, TMUX, "show-buffer", "-b", arg[0]))
					return
				}

				// 缓存列表
				for i, v := range kit.Split(m.Cmdx(cli.SYSTEM, TMUX, "list-buffers"), "\n", "\n", "\n") {
					ls := strings.SplitN(v, ": ", 3)
					m.Push(kit.MDB_NAME, ls[0])
					m.Push(kit.MDB_SIZE, ls[1])
					if i < 3 {
						m.Push(kit.MDB_TEXT, m.Cmdx(cli.SYSTEM, TMUX, "show-buffer", "-b", ls[0]))
					} else {
						m.Push(kit.MDB_TEXT, ls[2][1:len(ls[2])-1])
					}
				}
			}},
			TEXT: {Name: "text auto save text:textarea", Help: "文本", Action: map[string]*ice.Action{
				nfs.SAVE: {Name: "save", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
					if len(arg) > 0 && arg[0] != "" {
						m.Cmd(cli.SYSTEM, TMUX, "set-buffer", arg[0])
					}
					m.Cmdy(TEXT)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				text := m.Cmdx(cli.SYSTEM, TMUX, "show-buffer")
				m.EchoQRCode(text)
				m.EchoScript(text)
				m.Render("")
			}},
		},
	})
}
