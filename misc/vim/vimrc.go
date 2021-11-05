package vim

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const VIMRC = "vimrc"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		VIMRC: {Name: "vimrc", Help: "收藏夹", Action: ice.MergeAction(map[string]*ice.Action{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.PLUGIN, mdb.CREATE, VIMRC, m.Prefix(VIMRC))
				m.Cmd(mdb.RENDER, mdb.CREATE, VIMRC, m.Prefix(VIMRC))
				m.Cmd(mdb.PLUGIN, mdb.CREATE, VIM, m.Prefix(VIMRC))
				m.Cmd(mdb.RENDER, mdb.CREATE, VIM, m.Prefix(VIMRC))
				code.LoadPlug(m, VIMRC)
			}},
		}, code.PlugAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},
	}, Configs: map[string]*ice.Config{
		VIMRC: {Name: VIMRC, Help: "收藏夹", Value: kit.Data(
			code.PLUG, kit.Dict(
				code.SPLIT, kit.Dict("space", " \t", "operator", "{[(&.,;!|<>)]}"),
				code.PREFIX, kit.Dict("\"", "comment"),
				code.PREPARE, kit.Dict(
					code.KEYWORD, kit.Simple(
						"source", "finish",
						"set", "let", "end",
						"if", "else", "elseif", "endif",
						"for", "in", "continue", "break", "endfor",
						"try", "catch", "finally", "endtry",
						"call", "function", "return", "endfunction",

						"autocmd", "command", "execute",
						"nnoremap", "cnoremap", "inoremap",
						"colorscheme", "highlight", "syntax",
					),
					code.FUNCTION, kit.Simple(
						"has", "type", "empty",
						"exists", "executable",
					),
				),
			),
		)},
	}})
}
