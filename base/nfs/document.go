package nfs

import (
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const DOCUMENT = "document"

func init() {
	Index.MergeCommands(ice.Commands{
		DOCUMENT: {Name: "document index path auto", Help: "文档", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ice.COMMAND).Option(ice.MSG_DISPLAY, "")
				return
			}
			m.Search(arg[0], func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
				if p := DocumentPath(m); p != "" {
					if len(kit.Slice(arg, 0, 2)) == 1 {
						m.Cmdy(DIR, p)
					} else {
						m.Cmdy(CAT, arg[1])
					}
				}
			})
		}},
	})
}
func Document(m *ice.Message, p string, arg ...ice.Any) string {
	return kit.Renders(kit.Format(DocumentText(m, p), arg...), m)
}

var DocumentText = func(m *ice.Message, p string) string {
	return m.Cmdx(CAT, DocumentPath(m, path.Base(p)))
}
var DocumentPath = func(m *ice.Message, arg ...string) string {
	return path.Join(USR_LEARNING_PORTAL, m.PrefixKey(), path.Join(arg...))
}
