package wiki

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const BRIEF = "brief"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		BRIEF: {Name: "brief text", Help: "摘要", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_wiki_template(m, cmd, "", arg[0], arg[1:]...)
		}},
	}, Configs: map[string]*ice.Config{
		BRIEF: {Name: BRIEF, Help: "摘要", Value: kit.Data(
			nfs.TEMPLATE, `<p {{.OptionTemplate}}>{{.Option "text"}}</p>`,
		)},
	}})
}
