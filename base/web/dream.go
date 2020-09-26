package web

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/gdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"io/ioutil"
	"os"
	"path"
	"strings"
)

func _dream_list(m *ice.Message) {
	// 任务列表
	m.Cmdy(nfs.DIR, m.Conf(DREAM, "meta.path"), "time name").Table(func(index int, value map[string]string, head []string) {
		if m.Richs(SPACE, nil, value[kit.MDB_NAME], func(key string, value map[string]interface{}) {
			m.Push(kit.MDB_TYPE, value[kit.MDB_TYPE])
			m.Push(kit.MDB_STATUS, gdb.START)
		}) == nil {
			m.Push(kit.MDB_TYPE, "none")
			m.Push(kit.MDB_STATUS, gdb.STOP)
		}
		m.PushRender("action", "button", "start,stop,restart")
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
	os.MkdirAll(p, ice.MOD_DIR)

	if m.Option("repos") != "" {
		m.Cmd("web.code.git.repos", "create", "remote", m.Option("repos"), "path", p)
	}

	miss := path.Join(p, "etc/miss.sh")
	if _, e := os.Stat(miss); e != nil {
		m.Cmd(nfs.SAVE, miss, m.Conf(DREAM, "meta.miss"))
	}

	if b, e := ioutil.ReadFile(path.Join(p, m.Conf(gdb.SIGNAL, "meta.pid"))); e == nil {
		if s, e := os.Stat("/proc/" + string(b)); e == nil && s.IsDir() {
			m.Info("already exists %v", string(b))
			// 已经启动
			return
		}
	}

	if m.Richs(SPACE, nil, name, nil) == nil {
		// 启动任务
		m.Option(cli.CMD_DIR, p)
		m.Option(cli.CMD_STDERR, path.Join(p, m.Conf(DREAM, "meta.env.ctx_log")))
		m.Optionv(cli.CMD_ENV, kit.Simple(
			"ctx_dev", m.Conf(cli.RUNTIME, "conf.ctx_dev"),
			"PATH", kit.Path(path.Join(p, "bin"))+":"+os.Getenv("PATH"),
			"USER", cli.UserName,
			m.Confv(DREAM, "meta.env"),
		))
		m.Cmd(cli.DAEMON, m.Confv(DREAM, "meta.cmd"), "self", name)
	}
	m.Cmdy(nfs.DIR, p)
}

const DREAM = "dream"
const (
	DREAM_START = "dream.start"
	DREAM_STOP  = "dream.stop"
)

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			DREAM: {Name: "dream", Help: "梦想家", Value: kit.Data("path", "usr/local/work",
				"cmd", []interface{}{"ice.bin", SPACE, "connect"}, "env", kit.Dict(
					"ctx_log", "bin/boot.log", "ctx_mod", "ctx,log,gdb,ssh",
				),
				"miss", `#!/bin/bash
[ -f ~/.ish/plug.sh ] || [ -f ./.ish/plug.sh ] || git clone ${ISH_CONF_HUB_PROXY:="https://"}github.com/shylinux/intshell ./.ish
[ "$ISH_CONF_PRE" != "" ] || source ./.ish/plug.sh || source ~/.ish/plug.sh
require miss.sh

ish_miss_prepare_compile
ish_miss_prepare_install

# ish_miss_prepare_icebergs
# ish_miss_prepare_toolkits

`,
			)},
		},
		Commands: map[string]*ice.Command{
			DREAM: {Name: "dream [name [cmd...]] auto", Help: "梦想家", Meta: kit.Dict("detail", []interface{}{"启动", "停止"}), Action: map[string]*ice.Action{
				gdb.START: {Name: "start type name", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
					m.Option(kit.MDB_NAME, kit.Select(path.Base(m.Option(kit.SSH_REPOS)), m.Option(kit.MDB_NAME)))
					_dream_show(m, m.Option(kit.MDB_NAME))
				}},
				gdb.STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(SPACE, m.Option(kit.MDB_NAME), "exit", "1")
				}},
				gdb.RESTART: {Name: "restart", Help: "重启", Hand: func(m *ice.Message, arg ...string) {
					m.Cmd(SPACE, m.Option(kit.MDB_NAME), "exit", "1")
					m.Sleep("1s")
					_dream_show(m, m.Option(kit.MDB_NAME))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					_dream_list(m)
					return
				}
				if len(arg) > 1 {
					m.Cmdy(SPACE, arg[0], arg[1:])
					return
				}
				_dream_show(m, arg[0])
			}},
		},
	}, nil)
}
