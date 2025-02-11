package wx

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/core/code"
	kit "shylinux.com/x/toolkits"
)

const WXML = "wxml"

func init() {
	Index.MergeCommands(ice.Commands{
		WXML: {Actions: ice.MergeActions(code.PlugAction(code.PLUG, kit.Dict(
			code.INCLUDE, code.HTML,
			code.KEYWORD, kit.Dict(
				"page", code.KEYWORD,
				"view", code.KEYWORD,
				"text", code.KEYWORD,
				"image", code.KEYWORD,
				"picker", code.KEYWORD,
				"rich-text", code.KEYWORD,
				"template", code.KEYWORD,
				"import", code.KEYWORD,

				"class", code.FUNCTION,
				"type", code.FUNCTION,
				"open-type", code.FUNCTION,
				"size", code.FUNCTION,
				"mode", code.FUNCTION,
				"name", code.FUNCTION,
				"value", code.FUNCTION,
				"range", code.FUNCTION,
				"placeholder", code.FUNCTION,
				"is", code.FUNCTION,
				"data", code.FUNCTION,
				"nodes", code.FUNCTION,
				"wx:if", code.FUNCTION,
				"wx:elif", code.FUNCTION,
				"wx:else", code.FUNCTION,
				"wx:for", code.FUNCTION,
				"wx:for-item", code.FUNCTION,
				"wx:for-index", code.FUNCTION,
				"wx:key", code.FUNCTION,
				"bindtap", code.FUNCTION,
				"bindchange", code.FUNCTION,
				"bindconfirm", code.FUNCTION,
				"bindinput", code.FUNCTION,
				"bindblur", code.FUNCTION,
				"data-name", code.FUNCTION,
				"data-item", code.FUNCTION,
				"data-index", code.FUNCTION,
				"data-order", code.FUNCTION,
			),
		)), ice.Actions{
			code.TEMPLATE: {Hand: func(m *ice.Message, arg ...string) {
				m.Echo(`
<import src="../../app.wxml"/>
<view class="output">
	output
</view>
`)
			}},
		})},
	})
}
