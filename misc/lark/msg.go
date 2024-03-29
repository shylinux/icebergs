package lark

import (
	"encoding/json"
	"net/http"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

func _lark_get(m *ice.Message, appid string, arg ...ice.Any) (*ice.Message, ice.Any) {
	m.Option(web.SPIDE_HEADER, "Authorization", "Bearer "+m.Cmdx(APP, TOKEN, appid), web.ContentType, web.ApplicationJSON)
	msg := m.Cmd(web.SPIDE, LARK, http.MethodGet, arg)
	return msg, msg.Optionv(web.SPIDE_RES)
}
func _lark_post(m *ice.Message, appid string, arg ...ice.Any) *ice.Message {
	m.Option(web.SPIDE_HEADER, "Authorization", "Bearer "+m.Cmdx(APP, TOKEN, appid), web.ContentType, web.ApplicationJSON)
	return m.Cmd(web.SPIDE, LARK, arg)
}
func _lark_parse(m *ice.Message) {
	data := m.Optionv(ice.MSG_USERDATA)
	if data == nil {
		json.NewDecoder(m.R.Body).Decode(&data)
		m.Optionv(ice.MSG_USERDATA, data)

		switch d := data.(type) {
		case ice.Map:
			for k, v := range d {
				switch d := v.(type) {
				case ice.Map:
					for k, v := range d {
						m.Add(ice.MSG_OPTION, k, kit.Format(v))
					}
				default:
					for _, v := range kit.Simple(v) {
						m.Add(ice.MSG_OPTION, kit.Keys(MSG, k), kit.Format(v))
					}
				}
			}
		}
	}
	m.Debug("msg: %v", kit.Format(data))
}

const (
	APP_ID       = "app_id"
	SHIP_ID      = "ship_id"
	OPEN_ID      = "open_id"
	CHAT_ID      = "chat_id"
	USER_ID      = "user_id"
	OPEN_CHAT_ID = "open_chat_id"
	USER_OPEN_ID = "user_open_id"
)

const MSG = "msg"

func init() {
	Index.MergeCommands(ice.Commands{
		"_login": {Hand: func(m *ice.Message, arg ...string) {
			m.Option(ice.MSG_USERZONE, LARK)
		}},
		"/msg": {Name: "/msg", Help: "聊天消息", Hand: func(m *ice.Message, arg ...string) {
			data := m.Optionv(ice.MSG_USERDATA)
			if kit.Value(data, "action") != nil { // 卡片回调
				m.Cmd(MSG, "card")
				return
			}

			switch _lark_parse(m); m.Option("msg.type") {
			case "url_verification": // 绑定验证
				m.RenderResult(kit.Format(kit.Dict("challenge", m.Option("msg.challenge"))))

			case "event_callback": // 事件回调
				m.Cmd(EVENT, m.Option(mdb.TYPE))

			default: // 未知消息
				m.Cmd(DUTY, m.Option("msg.type"), kit.Formats(data))
			}
		}},
		MSG: {Name: "msg", Help: "聊天消息", Actions: ice.Actions{
			"location": {Hand: func(m *ice.Message, arg ...string) {}},
			"image":    {Hand: func(m *ice.Message, arg ...string) {}},
			"card": {Hand: func(m *ice.Message, arg ...string) {
				kit.For(kit.Value(m.Optionv(ice.MSG_USERDATA), "action.value"), func(k string, v string) { m.Option(k, v) })
				m.Cmdy(TALK, kit.Parse(nil, "", kit.Split(m.Option(ice.CMD))...))
				m.Cmd(SEND, m.Option(APP_ID), CHAT_ID, m.Option(OPEN_CHAT_ID), m.Option(wiki.TITLE)+lex.SP+m.Option(ice.CMD), m.Result())
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if m.Option(OPEN_CHAT_ID) == "" {
				m.Cmdy(DUTY, m.Option(mdb.TYPE), kit.Formats(m.Optionv(ice.MSG_USERDATA)))
			} else {
				m.Cmdy(TALK, strings.TrimSpace(m.Option("text_without_at_bot")))
			}
		}},
	})
}
