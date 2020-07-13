package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const SH = "sh"

func init() {
	Index.Register(&ice.Context{Name: SH, Help: "sh",
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmd(mdb.SEARCH, mdb.CREATE, SH, SH, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.SEARCH, mdb.CREATE, "man1", SH, c.Cap(ice.CTX_FOLLOW))
				m.Cmd(mdb.SEARCH, mdb.CREATE, "man8", SH, c.Cap(ice.CTX_FOLLOW))
			}},
			SH: {Name: SH, Help: "sh", Action: map[string]*ice.Action{
				mdb.SEARCH: {Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, m.Option("_path"))
					_c_find(m, kit.Select("main", arg, 1))
					_c_help(m, "1", kit.Select("main", arg, 1))
					_c_help(m, "8", kit.Select("main", arg, 1))
					_c_grep(m, kit.Select("main", arg, 1))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			}},
		},
	}, nil)
}
