package web

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ice.WEB_DREAM: {Name: "dream", Help: "梦想家", Value: kit.Data("path", "usr/local/work",
				// "cmd", []interface{}{ice.CLI_SYSTEM, "ice.sh", "start", ice.WEB_SPACE, "connect"},
				"cmd", []interface{}{ice.CLI_SYSTEM, "ice.bin", ice.WEB_SPACE, "connect"},
			)},
		},
		Commands: map[string]*ice.Command{
			ice.WEB_DREAM: {Name: "dream name auto", Help: "梦想家", Meta: kit.Dict(
				"exports", []string{"you", "name"}, "detail", []interface{}{"启动", "停止"},
			), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) > 1 && arg[0] == "action" {
					switch arg[1] {
					case "启动", "start":
						arg = []string{arg[4]}
					case "停止", "stop":
						m.Cmd(ice.WEB_SPACE, kit.Select(m.Option("name"), arg, 4), "exit", "1")
						m.Event(ice.DREAM_CLOSE, arg[4])
						return
					}
				}

				if len(arg) == 0 {
					// 任务列表
					m.Cmdy("nfs.dir", m.Conf(ice.WEB_DREAM, "meta.path"), "time name")
					m.Table(func(index int, value map[string]string, head []string) {
						if m.Richs(ice.WEB_SPACE, nil, value["name"], func(key string, value map[string]interface{}) {
							m.Push("type", value["type"])
							m.Push("status", "start")
						}) == nil {
							m.Push("type", "none")
							m.Push("status", "stop")
						}
					})
					m.Sort("name")
					m.Sort("status")
					return
				}

				// 规范命名
				if !strings.Contains(arg[0], "-") || !strings.HasPrefix(arg[0], "20") {
					arg[0] = m.Time("20060102-") + arg[0]
				}

				// 创建目录
				p := path.Join(m.Conf(ice.WEB_DREAM, "meta.path"), arg[0])
				os.MkdirAll(p, 0777)

				if b, e := ioutil.ReadFile(path.Join(p, m.Conf(ice.GDB_SIGNAL, "meta.pid"))); e == nil {
					if s, e := os.Stat("/proc/" + string(b)); e == nil && s.IsDir() {
						m.Info("already exists %v", string(b))
						return
					}
				}

				if m.Richs(ice.WEB_SPACE, nil, arg[0], nil) == nil {
					// 启动任务
					m.Option("cmd_dir", p)
					m.Option("cmd_type", "daemon")
					m.Optionv("cmd_env",
						"ctx_dev", m.Conf(ice.CLI_RUNTIME, "conf.ctx_dev"),
						"ctx_log", "boot.log", "ctx_mod", "ctx,log,gdb,ssh",
						"PATH", kit.Path(path.Join(p, "bin"))+":"+os.Getenv("PATH"),
					)
					m.Cmd(m.Confv(ice.WEB_DREAM, "meta.cmd"), "self", arg[0])
					time.Sleep(time.Second * 1)
					m.Event(ice.DREAM_START, arg...)
				}
				m.Cmdy("nfs.dir", p)
			}},
		},
	}, nil)
}
