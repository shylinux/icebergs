package mall

import (
	"github.com/shylinux/icebergs"
	_ "github.com/shylinux/icebergs/base"
	"github.com/shylinux/icebergs/base/web"
	"github.com/shylinux/toolkits"
	"os"
	"path"
	"strings"
)

var Index = &ice.Context{Name: "mall", Help: "商场模块",
	Caches: map[string]*ice.Cache{},
	Configs: map[string]*ice.Config{
		"miss": {Value: map[string]interface{}{
			"meta": map[string]interface{}{
				"path": "usr/local/work",
				"cmd":  []interface{}{"cli.system", "sh", "ice.sh", "start", "web.space", "connect"},
			},
			"list": map[string]interface{}{},
			"hash": map[string]interface{}{},
		}},
	},
	Commands: map[string]*ice.Command{
		"miss": {Name: "miss", Help: "任务", Meta: map[string]interface{}{
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
	},
}

func init() { web.Index.Register(Index, &web.WEB{}) }
