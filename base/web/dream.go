package web

import (
	ice "github.com/shylinux/icebergs"
	cli "github.com/shylinux/icebergs/base/cli"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
	"os"
	"path"
	"strings"
)

func _dream_list(m *ice.Message) {
	// 任务列表
	m.Cmdy("nfs.dir", m.Conf(DREAM, "meta.path"), "time name").Table(func(index int, value map[string]string, head []string) {
		if m.Richs(SPACE, nil, value[kit.MDB_NAME], func(key string, value map[string]interface{}) {
			m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
			m.Push(kit.MDB_STATUS, "start")
		}) == nil {
			m.Push(kit.MDB_TYPE, "none")
			m.Push(kit.MDB_STATUS, "stop")
		}
	})
	m.Sort(kit.MDB_NAME)
}
func _dream_show(m *ice.Message, name string) {
	// 规范命名
	if !strings.Contains(name, "-") || !strings.HasPrefix(name, "20") {
		name = m.Time("20060102-") + name
	}

	// 创建目录
	p := path.Join(m.Conf(DREAM, "meta.path"), name)
	os.MkdirAll(p, 0777)

	if b, e := ioutil.ReadFile(path.Join(p, m.Conf(ice.GDB_SIGNAL, "meta.pid"))); e == nil {
		if s, e := os.Stat("/proc/" + string(b)); e == nil && s.IsDir() {
			m.Info("already exists %v", string(b))
			return
		}
	}

	if m.Richs(SPACE, nil, name, nil) == nil {
		// 启动任务
		m.Option(cli.CMD_TYPE, "daemon")
		m.Option(cli.CMD_DIR, p)
		m.Optionv(cli.CMD_ENV,
			"ctx_dev", m.Conf(cli.RUNTIME, "conf.ctx_dev"),
			"ctx_log", "boot.log", "ctx_mod", "ctx,log,gdb,ssh",
			"PATH", kit.Path(path.Join(p, "bin"))+":"+os.Getenv("PATH"),
		)
		m.Cmd(m.Confv(DREAM, "meta.cmd"), "self", name)
	}
	m.Cmdy("nfs.dir", p)
}

const DREAM = "dream"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ice.WEB_DREAM: {Name: "dream", Help: "梦想家", Value: kit.Data("path", "usr/local/work",
				// "cmd", []interface{}{ice.CLI_SYSTEM, "ice.sh", "start", ice.WEB_SPACE, "connect"},
				"cmd", []interface{}{ice.CLI_SYSTEM, "ice.bin", ice.WEB_SPACE, "connect"},
			)},
		},
		Commands: map[string]*ice.Command{
			ice.WEB_DREAM: {Name: "dream [name] auto", Help: "梦想家", Meta: kit.Dict("detail", []interface{}{"启动", "停止"}), Action: map[string]*ice.Action{
				"start": {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					_dream_show(m, m.Option(kit.MDB_NAME))
				}},
				"stop": {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(SPACE, m.Option("name"), "exit", "1")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					_dream_list(m)
					return
				}
				_dream_show(m, arg[0])
			}},
		},
	}, nil)
}
