package lark

import (
	"encoding/json"
	"net/http"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

func _lark_get(m *ice.Message, appid string, arg ...interface{}) (*ice.Message, interface{}) {
	m.Option(web.SPIDE_HEADER, "Authorization", "Bearer "+m.Cmdx(APP, TOKEN, appid), web.ContentType, web.ContentJSON)
	msg := m.Cmd(web.SPIDE, LARK, http.MethodGet, arg)
	return msg, msg.Optionv(web.SPIDE_RES)
}
func _lark_post(m *ice.Message, appid string, arg ...interface{}) *ice.Message {
	m.Option(web.SPIDE_HEADER, "Authorization", "Bearer "+m.Cmdx(APP, TOKEN, appid), web.ContentType, web.ContentJSON)
	return m.Cmd(web.SPIDE, LARK, arg)
}
func _lark_parse(m *ice.Message) {
	data := m.Optionv(ice.MSG_USERDATA)
	if data == nil {
		json.NewDecoder(m.R.Body).Decode(&data)
		m.Optionv(ice.MSG_USERDATA, data)

		switch d := data.(type) {
		case map[string]interface{}:
			for k, v := range d {
				switch d := v.(type) {
				case map[string]interface{}:
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
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		web.WEB_LOGIN: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Option(ice.MSG_USERZONE, LARK)
		}},
		"/msg": {Name: "/msg", Help: "聊天消息", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
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
		MSG: {Name: "msg", Help: "聊天消息", Action: map[string]*ice.Action{
			"location": {Name: "", Help: "", Hand: func(m *ice.Message, arg ...string) {
			}},
			"image": {Name: "", Help: "", Hand: func(m *ice.Message, arg ...string) {
			}},
			"card": {Name: "", Help: "", Hand: func(m *ice.Message, arg ...string) {
				data := m.Optionv(ice.MSG_USERDATA)
				kit.Fetch(kit.Value(data, "action.value"), func(key string, value string) { m.Option(key, value) })
				m.Cmdy(TALK, kit.Parse(nil, "", kit.Split(m.Option(ice.CMD))...))
				m.Cmd(SEND, m.Option(APP_ID), CHAT_ID, m.Option(OPEN_CHAT_ID),
					m.Option(wiki.TITLE)+" "+m.Option(ice.CMD), m.Result())
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if m.Option(OPEN_CHAT_ID) == "" {
				m.Cmdy(DUTY, m.Option(mdb.TYPE), kit.Formats(m.Optionv(ice.MSG_USERDATA)))
			} else {
				m.Cmdy(TALK, strings.TrimSpace(m.Option("text_without_at_bot")))
			}
		}},
	}})
}
