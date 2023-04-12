package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	kit "shylinux.com/x/toolkits"
)

const FEEL = "feel"

func init() {
	Index.MergeCommands(ice.Commands{
		FEEL: {Name: "feel path auto prev next record1 record2 upload actions", Help: "影音媒体", Actions: ice.MergeActions(ice.Actions{
			"record1": {Help: "截图"}, "record2": {Help: "录屏"},
		}, WikiAction(ice.USR_LOCAL_IMAGE, "png|PNG|jpg|JPG|jpeg|mp4|m4v|mov|MOV|webm")), Hand: func(m *ice.Message, arg ...string) {
			_wiki_list(m, kit.Slice(arg, 0, 1)...)
			ctx.DisplayLocal(m, "")
		}},
	})
}
