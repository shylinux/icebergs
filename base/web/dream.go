package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
	"os"
	"path"
	"strings"
)

func _dream_list(m *ice.Message) {
	m.Cmdy(nfs.DIR, m.Conf(DREAM, kit.META_PATH), "time size name").Table(func(index int, value map[string]string, head []string) {
		if m.Richs(SPACE, nil, value[kit.MDB_NAME], func(key string, value map[string]interface{}) {
			m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
			m.Push(kit.MDB_STATUS, gdb.START)
			m.PushButton(gdb.STOP)
		}) == nil {
			m.Push(kit.MDB_TYPE, WORKER)
			m.Push(kit.MDB_STATUS, gdb.STOP)
			m.PushButton(gdb.START)
		}
	})
	m.Sort(kit.MDB_NAME, "str_r")
}
func _dream_show(m *ice.Message, name string) {
	// 规范命名
	if !strings.Contains(name, "-") || !strings.HasPrefix(name, "20") {
		name = m.Time("20060102-") + strings.ReplaceAll(name, "-", "_")
	}

	// 创建目录
	p := path.Join(m.Conf(DREAM, kit.META_PATH), name)
	os.MkdirAll(p, ice.MOD_DIR)

	// 下载代码
	if m.Option(kit.SSH_REPOS) != "" {
		m.Cmd("web.code.git.repos", mdb.CREATE, kit.SSH_REPOS, m.Option(kit.SSH_REPOS), kit.MDB_PATH, p)
	}

	if m.Option("template") != "" {
		for _, file := range []string{"etc/miss.sh", "go.mod", "src/main.go", "src/main.shy"} {
			if _, e := os.Stat(path.Join(p, file)); os.IsNotExist(e) {
				m.Cmdy(nfs.COPY, path.Join(p, file), path.Join(m.Option("template"), file))
				switch file {
				case "go.mod":
					kit.Rewrite(path.Join(p, file), func(line string) string {
						if strings.HasPrefix(line, "module") {
							m.Info("module %s", name)
							return "module " + name
						}
						return line
					})
				}
			}
		}

		go func() {
			m.Option(cli.CMD_DIR, p)
			m.Cmd(cli.SYSTEM, "bash", "-c", "source etc/miss.sh")
		}()
	}

	// 任务脚本
	miss := path.Join(p, "etc/miss.sh")
	if _, e := os.Stat(miss); os.IsNotExist(e) {
		m.Cmd(nfs.SAVE, miss, m.Conf(DREAM, "meta.miss"))
	}

	if b, e := ioutil.ReadFile(path.Join(p, m.Conf(gdb.SIGNAL, "meta.pid"))); e == nil {
		if s, e := os.Stat("/proc/" + string(b)); e == nil && s.IsDir() {
			m.Info("already exists %v", string(b))
			return // 已经启动
		}
	}

	if m.Richs(SPACE, nil, name, nil) == nil {
		m.Option(cli.CMD_DIR, p)
		m.Optionv(cli.CMD_ENV, kit.Simple(
			"ctx_dev", m.Conf(cli.RUNTIME, "conf.ctx_dev"),
			"PATH", kit.Path(path.Join(p, "bin"))+":"+kit.Path("bin")+":"+os.Getenv("PATH"),
			"USER", ice.Info.UserName, m.Confv(DREAM, "meta.env"),
		))
		// 启动任务
		m.Optionv(cli.CMD_STDERR, path.Join(p, m.Conf(DREAM, "meta.env.ctx_log")))
		m.Cmd(cli.DAEMON, m.Confv(DREAM, "meta.cmd"), "dev", "self", "name", name)
		m.Event(DREAM_CREATE, kit.MDB_TYPE, m.Option(kit.MDB_TYPE), kit.MDB_NAME, name)
		m.Sleep("1s")
	}
	m.Cmdy(nfs.DIR, p)
}

const (
	DREAM_CREATE = "dream.create"
	DREAM_START  = "dream.start"
	DREAM_STOP   = "dream.stop"
)
const DREAM = "dream"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			DREAM: {Name: DREAM, Help: "梦想家", Value: kit.Data(kit.MDB_PATH, "usr/local/work",
				"cmd", []interface{}{"ice.bin", SPACE, "connect"},
				"env", kit.Dict("ctx_log", "bin/boot.log"),
				"miss", `#!/bin/bash
[ -f ~/.ish/plug.sh ] || [ -f ./.ish/plug.sh ] || git clone ${ISH_CONF_HUB_PROXY:="https://"}github.com/shylinux/intshell ./.ish
[ "$ISH_CONF_PRE" != "" ] || source ./.ish/plug.sh || source ~/.ish/plug.sh
require miss.sh

ish_miss_prepare_compile
ish_miss_prepare_install

# ish_miss_prepare_icebergs
# ish_miss_prepare_toolkits

make
`,
			)},
		},
		Commands: map[string]*ice.Command{
			DREAM: {Name: "dream name path auto 启动", Help: "梦想家", Action: map[string]*ice.Action{
				gdb.START: {Name: "start name repos", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_NAME) == "self" {
						m.Option(kit.MDB_NAME, "")
					}
					_dream_show(m, m.Option(kit.MDB_NAME, kit.Select(path.Base(m.Option(kit.SSH_REPOS)), m.Option(kit.MDB_NAME))))
				}},
				gdb.STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option(kit.MDB_NAME), "exit", "0")
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_NAME:
						m.Cmdy(nfs.DIR, m.Conf(DREAM, kit.META_PATH), "name,time")
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					_dream_list(m)
					return
				}
				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(DREAM, kit.META_PATH), arg[0]))
				m.Cmdy(nfs.DIR, arg[1:])
			}},
		},
	})
}
