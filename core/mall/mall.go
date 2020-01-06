package mall

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"

	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

var Index = &ice.Context{Name: "mall", Help: "团队模块",
	Caches:  map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{},
	Commands: map[string]*ice.Command{
		ice.ICE_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmd(ice.WEB_SPIDE, "add", "12306", "https://kyfw.12306.cn")
		}},
		ice.ICE_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
		}},

		"_miss": {Name: "miss", Help: "任务", Meta: map[string]interface{}{
			"exports": []interface{}{"you", "name"},
			"detail":  []interface{}{"启动", "停止"},
		}, List: []interface{}{
			map[string]interface{}{"type": "text", "value": "", "name": "name"},
			map[string]interface{}{"type": "text", "value": "", "name": "type"},
			map[string]interface{}{"type": "button", "value": "创建", "action": "auto"},
		}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			if len(arg) > 1 {
				switch arg[1] {
				case "启动":
				case "停止":
					m.Cmd("web.space", arg[0], "exit", "1")
					return
				}
			}

			if len(arg) > 0 {
				if !strings.Contains(arg[0], "-") {
					arg[0] = m.Time("20060102-") + arg[0]
				}

				p := path.Join(m.Conf("miss", "meta.path"), arg[0])
				if _, e := os.Stat(p); e != nil {
					os.MkdirAll(p, 0777)
				}

				if !m.Confs("web.space", "hash."+arg[0]) {
					m.Option("cmd_dir", p)
					m.Option("cmd_type", "daemon")
					m.Cmd(m.Confv("miss", "meta.cmd"))
				}
			}

			m.Cmdy("nfs.dir", m.Conf("miss", "meta.path"), "", "time name")
			m.Table(func(index int, value map[string]string, head []string) {
				m.Push("status", kit.Select("stop", "start", m.Confs("web.space", "hash."+value["name"])))
			})
		}},
		"_task": {Name: "task", Help: "任务",
			Meta: map[string]interface{}{
				"remote": "true",
			},
			List: []interface{}{
				map[string]interface{}{"type": "select", "value": "create", "values": "create action cancel finish"},
				map[string]interface{}{"type": "text", "value": "", "name": "name"},
				map[string]interface{}{"type": "text", "value": "", "name": "text"},
				map[string]interface{}{"type": "button", "value": "创建"},
			}, Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
				switch arg[0] {
				case "create":
					id := m.Grow("web.chat.group", []string{kit.MDB_HASH, m.Option("sess.river"), "task"}, map[string]interface{}{
						"name":       arg[1],
						"text":       kit.Select("", arg, 2),
						"status":     "准备",
						"begin_time": m.Time(),
						"close_time": m.Time("3h"),
					})
					m.Log("info", "create task %v", id)
					m.Echo("%d", id)
				case "action":
				case "cancel":
				}
			}},
		"_process": {Name: "process", Help: "任务进度", Meta: map[string]interface{}{
			"remote": "true",
			"detail": []string{"编辑", "准备", "开始", "取消", "完成"},
		}, List: []interface{}{
			map[string]interface{}{"type": "text", "value": "0", "name": "offend"},
			map[string]interface{}{"type": "text", "value": "10", "name": "limit"},
			map[string]interface{}{"type": "text", "value": "", "name": "match"},
			map[string]interface{}{"type": "text", "value": "", "name": "value"},
			map[string]interface{}{"type": "button", "value": "查看", "action": "auto"},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {

			switch arg[1] {
			case "modify":
				prefix := []string{"mdb.update", "web.chat.group", "hash." + m.Option("sess.river") + ".task", arg[0], arg[2], arg[3]}
				m.Cmd(prefix)
				arg = arg[4:]

			case "准备", "开始", "取消", "完成":
				msg := m.Cmd("mdb.select", "web.chat.group", "hash."+m.Option("sess.river")+".task", arg[0])
				if msg.Append("status") == arg[1] {
					arg = arg[4:]
					break
				}
				prefix := []string{"mdb.update", "web.chat.group", "hash." + m.Option("sess.river") + ".task", arg[0], "status", arg[1]}
				status := map[string][]string{
					"准备->开始": []string{"begin_time", m.Time(), "close_time", m.Time("3h")},
					"准备->取消": []string{"begin_time", m.Time(), "close_time", m.Time()},
					"准备->完成": []string{"begin_time", m.Time(), "close_time", m.Time()},

					"开始->准备": []string{"begin_time", m.Time(), "close_time", m.Time("3h")},
					"开始->取消": []string{"close_time", m.Time()},
					"开始->完成": []string{"close_time", m.Time()},
				}[msg.Append("status")+"->"+arg[1]]
				suffix := []string{"extra.list.-2.time", m.Time(), "extra.list.-3.status", arg[1]}

				if len(status) > 0 {
					m.Cmd(prefix, status, suffix)
				}
				arg = arg[4:]
			}

			m.Cmdy("mdb.select", "web.chat.group", "hash."+m.Option("sess.river")+".task", "0",
				kit.Select("0", arg, 0), kit.Select("10", arg, 1), kit.Select("", arg, 2), kit.Select("", arg, 3),
				kit.Select("id status begin_time close_time name text", arg, 4))
		}},
		"12306": &ice.Command{Name: "12306", Help: "12306", Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			date := time.Now().Add(time.Hour * 24).Format("2006-01-02")
			if len(arg) > 0 {
				date, arg = arg[0], arg[1:]
			}
			to := "QFK"
			if len(arg) > 0 {
				to, arg = arg[0], arg[1:]
			}
			from := "BJP"
			if len(arg) > 0 {
				from, arg = arg[0], arg[1:]
			}
			m.Echo("%s->%s %s\n", from, to, date)

			m.Cmd(ice.WEB_SPIDE, "12306", "GET", fmt.Sprintf("/otn/leftTicket/init?linktypeid=dc&fs=北京,BJP&ts=曲阜,QFK&date=%s&flag=N,N,Y", date))
			m.Cmd(ice.WEB_SPIDE, "12306", "GET", fmt.Sprintf("/otn/leftTicket/queryZ?leftTicketDTO.train_date=%s&leftTicketDTO.from_station=%s&leftTicketDTO.to_station=%s&purpose_codes=ADULT", date, from, to)).Table(func(index int, value map[string]string, head []string) {
				kit.Fetch(kit.Value(kit.UnMarshal(value["data"]), "result"), func(index int, value string) {
					fields := strings.Split(value, "|")
					m.Info("once %d len(%d) %v", index, len(fields), fields)
					m.Push("车次--", fields[3])
					m.Push("出发----", fields[8])
					m.Push("到站----", fields[9])
					m.Push("时长----", fields[10])
					m.Push("二等座", fields[30])
					m.Push("一等座", fields[31])
				})
			})
			return
		}},
	},
}

func init() { web.Index.Register(Index, &web.Frame{}) }
