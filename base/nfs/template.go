package nfs

import (
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const TEMPLATE = "template"

func init() {
	Index.MergeCommands(ice.Commands{
		TEMPLATE: {Name: "template index path auto", Help: "模板", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				ice.AddRender(ice.RENDER_TEMPLATE, func(m *ice.Message, args ...ice.Any) string {
					return Template(m, kit.Format(args[0]), args[1:]...)
				})
			}},
		}), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ice.COMMAND).Action("filter:text").Option(ice.MSG_DISPLAY, "")
				return
			}
			m.Search(arg[0], func(p *ice.Context, c *ice.Context, key string, cmd *ice.Command) {
				if p := TemplatePath(m); p != "" {
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

func init() { ice.Info.Template = Template }
func Template(m *ice.Message, p string, data ...ice.Any) string {
	if len(data) == 0 {
		return kit.Renders(TemplateText(m, p), m)
	}
	return kit.Renders(TemplateText(m, p), data[0])
}

var TemplateText = func(m *ice.Message, p string) string {
	return m.Cmdx(CAT, kit.Select(TemplatePath(m, path.Base(p)), m.Option("_template")))
}
var TemplatePath = func(m *ice.Message, arg ...string) string {
	if p := path.Join(ice.SRC_TEMPLATE, m.PrefixKey(), path.Join(arg...)); Exists(m, p) {
		return p
	} else {
		return p
	}
}
