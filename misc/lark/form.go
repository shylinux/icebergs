package lark

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/aaa"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const FORM = "form"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		FORM: {Name: "form [chat_id|open_id|user_id|email] target title text [confirm|value|url arg...]...", Help: "消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			var form = kit.Dict(CONTENT, kit.Dict())
			switch arg[0] {
			case CHAT_ID, OPEN_ID, USER_ID, aaa.EMAIL:
				form[arg[0]], arg = arg[1], arg[2:]
			default:
				form[CHAT_ID], arg = arg[0], arg[1:]
			}

			elements := []interface{}{}
			elements = append(elements, kit.Dict(
				"tag", "div", "text", kit.Dict(
					"tag", "plain_text", CONTENT, kit.Select(" ", arg[1]),
				),
			))

			actions := []interface{}{}
			for i := 2; i < len(arg); i++ {
				button := kit.Dict(
					"tag", "button", "type", "default", "text", kit.Dict(
						"tag", "plain_text", CONTENT, kit.Select(" ", arg[i]),
					),
				)

				content := arg[i]
				switch arg[i+1] {
				case "confirm":
					button[arg[i+1]], i = kit.Dict(
						"title", kit.Dict("tag", "lark_md", CONTENT, arg[i+2]),
						"text", kit.Dict("tag", "lark_md", CONTENT, arg[i+3]),
					), i+3
				case "value":
					button[arg[i+1]], i = kit.Dict(arg[i+2], arg[i+3]), i+3
				case "url":
					button[arg[i+1]], i = arg[i+2], i+2
				default:
					button["value"], i = kit.Dict(
						ice.MSG_RIVER, m.Option(ice.MSG_RIVER),
						ice.MSG_STORM, m.Option(ice.MSG_STORM),
						arg[i+1], arg[i+2],
					), i+2
				}
				kit.Value(button, "value.content", content)
				kit.Value(button, "value.open_chat_id", m.Option(OPEN_CHAT_ID))
				kit.Value(button, "value.description", arg[1])
				kit.Value(button, "value.title", arg[0])

				actions = append(actions, button)
			}
			elements = append(elements, kit.Dict("tag", "action", "actions", actions))

			kit.Value(form, "msg_type", "interactive")
			kit.Value(form, "card", kit.Dict(
				"config", kit.Dict("wide_screen_mode", true),
				"header", kit.Dict(
					"title", kit.Dict("tag", "lark_md", CONTENT, arg[0]),
				),
				"elements", elements,
			))

			msg := _lark_post(m, m.Option(APP_ID), "/open-apis/message/v4/send/", web.SPIDE_DATA, kit.Formats(form))
			m.Debug("%v", msg.Optionv(web.SPIDE_RES))
		}},
	}})
}
