package wiki

import (
	ice "shylinux.com/x/icebergs"
)

const FEEL = "feel"

func init() {
	Index.MergeCommands(ice.Commands{
		FEEL: {Name: "feel path auto record1 record upload prev next actions", Help: "影音媒体", Actions: ice.MergeActions(ice.Actions{
			"record1": {Name: "record1", Help: "截图", Hand: func(m *ice.Message, arg ...string) {}},
			"record":  {Name: "record", Help: "录屏", Hand: func(m *ice.Message, arg ...string) {}},
		}, WikiAction(ice.USR_LOCAL_IMAGE, "png|PNG|jpg|JPG|jpeg|mp4|m4v|MOV|webm")), Hand: func(m *ice.Message, arg ...string) {
			_wiki_list(m, arg...)
		}},
	})
}
