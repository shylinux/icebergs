package wx

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

const TEMPLATE = "template"

func init() {
	Index.MergeCommands(ice.Commands{
		TEMPLATE: {Name: "template access template_id openid auto", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				m.Cmdy(ACCESS)
			} else if m.Option(ACCESS, arg[0]); len(arg) == 1 {
				res := SpideGet(m, "template/get_all_private_template")
				kit.For(kit.Value(res, "template_list"), func(value ice.Map) {
					m.Push("", value)
				})
				m.Cut("template_id,title,content,example")
			} else if len(arg) > 4 {
				args := []ice.Any{"template_id", arg[1], "touser", arg[2], "url", arg[3]}
				kit.For(arg[4:], func(k, v string) { args = append(args, kit.Keys("data", k, "value"), v) })
				SpidePost(m, "message/template/send", args...)
			}
		}},
	})
}
