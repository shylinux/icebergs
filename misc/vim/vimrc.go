package vim

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/core/code"
)

const VIMRC = "vimrc"

func init() {
	Index.MergeCommands(ice.Commands{
		VIM:   {Name: "vim", Actions: code.PlugAction()},
		VIMRC: {Name: "vimrc", Actions: code.PlugAction()},
	})
}
