package wiki

import ice "shylinux.com/x/icebergs"

const IFRAME = "iframe"

func init() {
	Index.MergeCommands(ice.Commands{
		IFRAME: {Name: "iframe link", Hand: func(m *ice.Message, arg ...string) {
			_wiki_template(m, "", "", arg[0], arg[1:]...)
		}},
	})
}
