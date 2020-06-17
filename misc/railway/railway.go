package railway

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/core/wiki"
	"github.com/shylinux/toolkits"
)

var Index = &ice.Context{Name: "railway", Help: "railway",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"railway": {Name: "railway", Help: "12306", Value: kit.Data("site", "https://kyfw.12306.cn")},
	},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Load()
			web.SpideCreate(m, "12306", m.Conf("railway", "meta.site"))
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Save("railway")
		}},

		"railway": &ice.Command{Name: "railway", Help: "12306", List: kit.List(
			kit.MDB_INPUT, "text", "name", "date", "figure", "date",
			kit.MDB_INPUT, "text", "name", "from", "value", "北京", "figure", "city",
			kit.MDB_INPUT, "text", "name", "to", "value", "曲阜", "figure", "city",
			kit.MDB_INPUT, "button", "name", "查询",
		), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if !m.Confs("railway", "meta.place") {
				list := strings.Split(strings.TrimPrefix(m.Cmdx(ice.WEB_SPIDE, "12306", "raw", "GET", "/otn/resources/js/framework/station_name.js?station_version=1.9090"), "var statuion_names ='"), "|")
				for i := 0; i < len(list)-5; i += 5 {
					m.Conf("railway", kit.Keys("meta.place", list[i+1]), list[i+2])
				}
			}

			date := strings.Split(m.Time("24h"), " ")[0]
			if len(arg) > 0 {
				date, arg = arg[0], arg[1:]
			}
			date = strings.Split(date, " ")[0]
			from := "北京"
			if len(arg) > 0 {
				from, arg = arg[0], arg[1:]
			}
			from_code := m.Conf("railway", kit.Keys("meta.place", from))
			to := "曲阜"
			if len(arg) > 0 {
				to, arg = arg[0], arg[1:]
			}
			to_code := m.Conf("railway", kit.Keys("meta.place", to))

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
		"passcode": &ice.Command{Name: "passcode", Help: "passcode", Meta: kit.Dict("active", "mall/input"), Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			prefix := []string{ice.WEB_SPIDE, "12306"}
			if len(arg) == 0 {
				m.Cmd(prefix, "raw", "/passport/web/auth/uamtk-static", "form", "appid", "otn")
				m.Cmd(prefix, "raw", "GET", "/otn/HttpZF/GetJS")
				m.Cmd(prefix, "raw", "/otn/login/conf")

				m.Cmdy(prefix, "GET", fmt.Sprintf("/passport/captcha/captcha-image64?login_site=E&module=login&rand=sjrand"))
				return
			}

			switch arg[0] {
			case "check":
				m.Cmdy(prefix, "GET", fmt.Sprintf("/passport/captcha/captcha-check?login_site=E&rand=sjrand&answer=%s", arg[1]))
			case "login":
				m.Cmdy(prefix, "raw", "/passport/web/login", "form", "username", arg[1], "password", arg[2], "answer", arg[3], "appid", "otn")
			}
		}},
	},
}

func init() { wiki.Index.Register(Index, nil) }
