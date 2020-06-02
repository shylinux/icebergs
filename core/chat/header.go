package chat

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
)

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/header": {Name: "/header", Help: "标题栏", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			switch kit.Select("", arg, 0) {
			case "check":
				if m.Option(ice.MSG_USERNAME) != "" {
					m.Echo(m.Option(ice.MSG_USERNAME))
				}
			case "login":
				if m.Option(ice.MSG_USERNAME) != "" {
					m.Render(m.Option(ice.MSG_USERNAME))
				}
			default:
				m.Echo(m.Conf(ice.WEB_SERVE, "meta.title"))
			}
		}},
	}}, nil)
}
