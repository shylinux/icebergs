package wx

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/core/chat"
	"shylinux.com/x/icebergs/core/wiki"
	kit "shylinux.com/x/toolkits"
)

func _wx_parse(m *ice.Message) {
	data := struct {
		FromUserName string
		ToUserName   string
		CreateTime   int64
		MsgID        int64
		Event        string
		MsgType      string
		Content      string

		Location_X float64
		Location_Y float64
		Scale      string
		Label      string

		Title       string
		Description string
		Url         string

		PicUrl string
	}{}
	buf, _ := ioutil.ReadAll(m.R.Body)
	m.Debug("buf: %+v", string(buf))
	xml.NewDecoder(bytes.NewBuffer(buf)).Decode(&data)
	m.Debug("data: %+v", data)

	m.Option("FromUserName", data.FromUserName)
	m.Option("ToUserName", data.ToUserName)
	m.Option("CreateTime", data.CreateTime)
	m.Option("MsgID", data.MsgID)

	m.Option("Event", data.Event)
	m.Option("MsgType", data.MsgType)
	m.Option("Content", data.Content)

	m.Option("LocationX", kit.Int(data.Location_X*100000))
	m.Option("LocationY", kit.Int(data.Location_Y*100000))
	m.Option("Scale", data.Scale)
	m.Option("Label", data.Label)

	m.Option("Title", data.Title)
	m.Option("Description", data.Description)
	m.Option("URL", data.Url)

	m.Option("URL", data.PicUrl)
}

const LOGIN = "login"

func init() {
	Index.MergeCommands(ice.Commands{
		"/login/": {Name: "/login/", Help: "认证", Hand: func(m *ice.Message, arg ...string) {
			if m.Cmdx(ACCESS, CHECK) == "" {
				return // 验签失败
			}

			// 解析数据
			_wx_parse(m)

			// 用户登录
			m.Option(ice.MSG_USERZONE, WX)
			aaa.UserLogin(m, m.Append("FromUserName"), "")

			switch m.Option("MsgType") {
			case EVENT: // 事件
				m.Cmdy(EVENT, m.Option("Event"))

			case chat.LOCATION: // 打卡
				m.Cmdy(chat.LOCATION, mdb.CREATE, mdb.TYPE, "", mdb.NAME, m.Option("Label"), mdb.TEXT, m.Option("Label"),
					"latitude", m.Option("LocationX"), "longitude", m.Option("LocationY"), "scale", m.Option("Scale"),
				)
			case mdb.LINK: // 打卡
				m.Cmdy(chat.SCAN, mdb.CREATE, mdb.TYPE, mdb.LINK, mdb.NAME, m.Option("Title"), mdb.TEXT, m.Option("URL"))

			case "image": // 文本
				m.Cmdy(chat.SCAN, mdb.CREATE, mdb.TYPE, wiki.IMAGE, mdb.NAME, m.Option("Title"), mdb.TEXT, m.Option("URL"))

			case TEXT: // 文本
				if cmds := kit.Split(m.Option("Content")); m.Right(cmds) {
					m.Cmdy(TEXT, cmds)
					break
				}
				m.Cmdy(MENU, "home")
			}
		}},
		LOGIN: {Name: "login", Help: "登录", Actions: ice.Actions{
			mdb.CREATE: {Name: "create appid appmm token", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Conf(ACCESS, kit.Keym(APPID), m.Option(APPID))
				m.Conf(ACCESS, kit.Keym(APPMM), m.Option(APPMM))
				m.Conf(ACCESS, kit.Keym(TOKEN), m.Option(TOKEN))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
		}},
	})
}
