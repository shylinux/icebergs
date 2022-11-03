package wiki

import (
	ice "shylinux.com/x/icebergs"
)

const BRIEF = "brief"

func init() {
	Index.MergeCommands(ice.Commands{
		BRIEF: {Name: "brief text", Help: "摘要", Actions: WordAction(
			`<p {{.OptionTemplate}}>{{.Option "text"}}</p>`,
		), Hand: func(m *ice.Message, arg ...string) { _wiki_template(m, "", arg[0], arg[1:]...) }},
	})
}
