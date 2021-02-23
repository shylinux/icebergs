package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	"github.com/shylinux/icebergs/base/tcp"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
	"os"
	"path"
	"strings"
)

func _dream_list(m *ice.Message) {
	m.Cmdy(nfs.DIR, m.Conf(DREAM, kit.META_PATH), "time,size,name").Table(func(index int, value map[string]string, head []string) {
		if m.Richs(SPACE, nil, value[kit.MDB_NAME], func(key string, value map[string]interface{}) {
			m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
			m.Push(kit.MDB_STATUS, tcp.START)
			m.PushButton(tcp.STOP)
		}) == nil {
			m.Push(kit.MDB_TYPE, WORKER)
			m.Push(kit.MDB_STATUS, tcp.STOP)
			m.PushButton(tcp.START)
		}
	})
	m.SortStrR(kit.MDB_TIME)
}
func _dream_show(m *ice.Message, name string) {
	if !strings.Contains(name, "-") || !strings.HasPrefix(name, "20") {
		name = m.Time("20060102-") + strings.ReplaceAll(name, "-", "_")
	}
	m.Option("name", name)

	// 任务目录
	p := path.Join(m.Conf(DREAM, kit.META_PATH), name)
	if m.Option(kit.SSH_REPOS) != "" { // 下载源码
		m.Cmd("web.code.git.repos", mdb.CREATE, kit.SSH_REPOS, m.Option(kit.SSH_REPOS), kit.MDB_PATH, p)
	} else { // 创建目录
		os.MkdirAll(p, ice.MOD_DIR)
	}

	// 任务模板
	if m.Option("template") != "" {
		for _, file := range []string{"etc/miss.sh", "src/main.shy", "src/main.go", "go.mod", "Makefile"} {
			if _, e := os.Stat(path.Join(p, file)); os.IsNotExist(e) {
				switch m.Cmdy(nfs.COPY, path.Join(p, file), path.Join(m.Option("template"), file)); file {
				case "go.mod":
					kit.Rewrite(path.Join(p, file), func(line string) string {
						return kit.Select(line, "module "+name, strings.HasPrefix(line, "module"))
					})
				}
			}
		}
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
			// "ctx_dev", m.Conf(cli.RUNTIME, "conf.ctx_dev"),
			"ctx_dev", "http://:"+m.Cmd(SERVE).Append(tcp.PORT),
			"PATH", kit.Path(path.Join(p, "bin"))+":"+kit.Path("bin")+":"+os.Getenv("PATH"),
			"USER", ice.Info.UserName, m.Confv(DREAM, "meta.env"),
		))
		// 启动任务
		kit.Path(os.Args[0])

		m.Optionv(cli.CMD_STDERR, path.Join(p, m.Conf(DREAM, "meta.env.ctx_log")))
		m.Cmd(cli.DAEMON, m.Confv(DREAM, "meta.cmd"), SPIDE_DEV, SPIDE_DEV, kit.MDB_NAME, name)
		m.Event(DREAM_CREATE, kit.MDB_TYPE, m.Option(kit.MDB_TYPE), kit.MDB_NAME, name)
		m.Sleep(ice.MOD_TICK)
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
		Commands: map[string]*ice.Command{
			DREAM: {Name: "dream name path auto start create", Help: "梦想家", Action: map[string]*ice.Action{
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					switch arg[0] {
					case kit.MDB_NAME:
						m.Cmdy(nfs.DIR, m.Conf(DREAM, kit.META_PATH), "name,time")
					case kit.SSH_TEMPLATE:
						m.Cmdy(nfs.DIR, m.Conf(DREAM, kit.META_PATH), "path,size,time")
						m.SortStrR(kit.MDB_PATH)
					}
				}},
				mdb.CREATE: {Name: "create main=src/main.go@key name=hi@key from=usr/icebergs/misc/bash/bash.go@key", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option(ROUTE), "web.code.autogen", mdb.CREATE, arg)
					m.Option(ice.MSG_PROCESS, ice.PROCESS_INNER)
				}},
				tcp.START: {Name: "start name repos", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_NAME) == SPIDE_SELF {
						m.Option(kit.MDB_NAME, "")
					}
					_dream_show(m, m.Option(kit.MDB_NAME, kit.Select(path.Base(m.Option(kit.SSH_REPOS)), m.Option(kit.MDB_NAME))))
				}},
				tcp.STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(SPACE, m.Option(kit.MDB_NAME), "exit", "0")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					_dream_list(m)
					return
				}

				m.Option(nfs.DIR_ROOT, path.Join(m.Conf(DREAM, kit.META_PATH), arg[0]))
				if len(arg) == 1 || strings.HasSuffix(arg[1], "/") {
					m.Cmdy(nfs.DIR, arg[1:])
				} else {
					m.Cmdy(nfs.CAT, arg[1:])
				}
			}},
		},
		Configs: map[string]*ice.Config{
			DREAM: {Name: DREAM, Help: "梦想家", Value: kit.Data(kit.MDB_PATH, "usr/local/work",
				"cmd", []interface{}{"ice.bin", SPACE, tcp.DIAL},
				"env", kit.Dict("ctx_log", "bin/boot.log"),
				"miss", `#!/bin/bash
[ -f ~/.ish/plug.sh ] || [ -f $PWD/.ish/plug.sh ] || git clone ${ISH_CONF_HUB_PROXY:="https://"}github.com/shylinux/intshell $PWD/.ish
[ "$ISH_CONF_PRE" != "" ] || source $PWD/.ish/plug.sh || source ~/.ish/plug.sh
require miss.sh

ish_miss_prepare_develop
ish_miss_prepare_compile
ish_miss_prepare_install

# ish_miss_prepare_volcanos
# ish_miss_prepare_learning
# ish_miss_prepare_icebergs
# ish_miss_prepare_toolkits

make
`,
			)},
		},
	})
}
