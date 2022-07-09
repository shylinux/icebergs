package lark

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _send_text(m *ice.Message, form ice.Map, arg ...string) bool {
	switch len(arg) {
	case 0:
	case 1:
		kit.Value(form, "msg_type", "text")
		kit.Value(form, "content.text", arg[0])
		if strings.TrimSpace(arg[0]) == "" {
			return false
		}
	default:
		if len(arg) == 2 && strings.TrimSpace(arg[1]) == "" {
			return false
		}
		content := []ice.Any{}
		line := []ice.Any{}
		for _, v := range arg[1:] {
			if v == "\n" {
				content, line = append(content, line), []ice.Any{}
				continue
			}
			line = append(line, ice.Map{"tag": "text", "text": v + " "})
		}
		content = append(content, line)

		kit.Value(form, "msg_type", "post")
		kit.Value(form, "content.post", ice.Map{
			"zh_cn": ice.Map{"title": arg[0], CONTENT: content},
		})
	}
	return true
}

const (
	CONTENT = "content"
)
const SEND = "send"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		SEND: {Name: "send appid [chat_id|open_id|user_id|email] target [title] text", Help: "消息", Hand: func(m *ice.Message, arg ...string) {
			form := kit.Dict(CONTENT, kit.Dict())
			appid, arg := arg[0], arg[1:]
			switch arg[0] {
			case CHAT_ID, OPEN_ID, USER_ID, aaa.EMAIL:
				form[arg[0]], arg = arg[1], arg[2:]
			default:
				form[CHAT_ID], arg = arg[0], arg[1:]
			}

			if _send_text(m, form, arg...) {
				msg := _lark_post(m, appid, "/open-apis/message/v4/send/", web.SPIDE_DATA, kit.Format(form))
				m.Push(mdb.TIME, m.Time())
				m.Push("message_id", msg.Append("data.message_id"))
			}
		}},
	}})
}
