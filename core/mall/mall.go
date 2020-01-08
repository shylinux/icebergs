package mall

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"fmt"
	"strings"
	"time"
)

var Index = &ice.Context{Name: "mall", Help: "团队模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"railway": {Name: "railway", Help: "12306", Value: kit.Data()},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "load", "mall.json")
			if m.Richs(ice.WEB_SPIDE, nil, "12306", nil) == nil {
				m.Cmd(ice.WEB_SPIDE, "add", "12306", "https://kyfw.12306.cn")
			}
			if !m.Confs("railway", "meta.site") {
				list := strings.Split(strings.TrimPrefix(m.Cmdx(ice.WEB_SPIDE, "12306", "raw", "GET", "/otn/resources/js/framework/station_name.js?station_version=1.9090"), "var statuion_names ='"), "|")
				for i := 0; i < len(list)-5; i += 5 {
					m.Conf("railway", kit.Keys("meta.site", list[i+1]), list[i+2])
				}
			}
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.CTX_CONFIG, "save", "mall.json", "web.mall.railway")
		}},

		"passcode": &ice.Command{Name: "passcode", Help: "passcode", Meta: kit.Dict(
			"display", "mall/image",
		), List: kit.List(
			kit.MDB_INPUT, "text", "name", "账号",
			kit.MDB_INPUT, "text", "name", "密码",
			kit.MDB_INPUT, "button", "name", "登录", "display", "mall/input",
			kit.MDB_INPUT, "button", "name", "刷新", "display", "mall/input",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) == 0 {
				m.Cmd(ice.WEB_SPIDE, "12306", "raw", "/passport/web/auth/uamtk-static", "form", "appid", "otn")
				m.Cmd(ice.WEB_SPIDE, "12306", "raw", "GET", "/otn/HttpZF/GetJS")
				m.Cmd(ice.WEB_SPIDE, "12306", "raw", "/otn/login/conf")

				m.Cmdy(ice.WEB_SPIDE, "12306", "GET", fmt.Sprintf("/passport/captcha/captcha-image64?login_site=E&module=login&rand=sjrand"))
				return
			}

			switch arg[0] {
			case "check":
				m.Cmdy(ice.WEB_SPIDE, "12306", "GET", fmt.Sprintf("/passport/captcha/captcha-check?answer=%s&rand=sjrand&login_site=E", arg[1]))
			case "login":
				m.Cmdy(ice.WEB_SPIDE, "12306", "raw", "/passport/web/login", "form", "username", arg[1], "password", arg[2], "answer", arg[3], "appid", "otn")
			}
		}},
		"railway": &ice.Command{Name: "railway", Help: "12306", List: kit.List(
			kit.MDB_INPUT, "text", "name", "date", "value", "2020-01-22",
			kit.MDB_INPUT, "text", "name", "from", "value", "北京",
			kit.MDB_INPUT, "text", "name", "to", "value", "曲阜",
			kit.MDB_INPUT, "button", "name", "查询",
		),
			Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				date := time.Now().Add(time.Hour * 24).Format("2006-01-02")
				if len(arg) > 0 {
					date, arg = arg[0], arg[1:]
				}
				from := "北京"
				if len(arg) > 0 {
					from, arg = arg[0], arg[1:]
				}
				from_code := m.Conf("railway", kit.Keys("meta.site", from))
				to := "曲阜"
				if len(arg) > 0 {
					to, arg = arg[0], arg[1:]
				}
				to_code := m.Conf("railway", kit.Keys("meta.site", to))

				m.Echo("%s->%s %s\n", from, to, date)

				if len(arg) > 0 {
					m.Cmdy(ice.WEB_SPIDE, "12306", "raw", "GET", fmt.Sprintf("/otn/czxx/queryByTrainNo?train_no=%s&from_station_telecode=%s&to_station_telecode=%s&depart_date=%s",
						arg[0], from_code, to_code, date))
					return
				}

				m.Cmd(ice.WEB_SPIDE, "12306", "GET", fmt.Sprintf("/otn/leftTicket/init?linktypeid=dc&fs=%s,%s&ts=%s,%s&date=%s&flag=N,N,Y",
					from, from_code, to, to_code, date))
				m.Cmd(ice.WEB_SPIDE, "12306", "GET", fmt.Sprintf("/otn/leftTicket/queryZ?leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=ADULT",
					date, from_code, to_code)).Table(func(index int, value map[string]string, head []string) {
					kit.Fetch(kit.Value(kit.UnMarshal(value["data"]), "result"), func(index int, value string) {
						fields := strings.Split(value, "|")
						m.Push("车次", fields[3])
						m.Push("出发", fields[8])
						m.Push("到站", fields[9])
						m.Push("时长", fields[10])
						m.Push("二等座", fields[30])
						m.Push("一等座", fields[31])
					})
				})
			}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
