package vim

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/core/code"
)

const VIMRC = "vimrc"

func init() {
	Index.MergeCommands(ice.Commands{
		VIM:   {Actions: code.PlugAction()},
		VIMRC: {Actions: code.PlugAction()},
	})
}
