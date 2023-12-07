package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const WXSS = "wxss"

func init() {
	Index.MergeCommands(ice.Commands{
		WXSS: {Actions: code.PlugAction(code.PLUG, kit.Dict(
			code.INCLUDE, code.CSS,
			code.KEYWORD, kit.Dict(
				"page", code.KEYWORD,
				"view", code.KEYWORD,
				"text", code.KEYWORD,
				"image", code.KEYWORD,
				"picker", code.KEYWORD,
				"rich-text", code.KEYWORD,
			),
		))},
	})
}
