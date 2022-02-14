package web

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _dream_list(m *ice.Message) *ice.Message {
	return m.Cmdy(nfs.DIR, m.Config(nfs.PATH), "time,size,name").Table(func(index int, value map[string]string, head []string) {
		if m.Richs(SPACE, nil, value[mdb.NAME], func(key string, val map[string]interface{}) {
			m.Push(mdb.TYPE, val[mdb.TYPE])
			m.Push(cli.STATUS, cli.START)
			m.PushButton(cli.STOP)
			m.PushAnchor(strings.Split(kit.MergePOD(m.Option(ice.MSG_USERWEB), value[mdb.NAME]), "?")[0])
		}) == nil {
			m.Push(mdb.TYPE, WORKER)
			m.Push(cli.STATUS, cli.STOP)
			m.PushButton(cli.START)
			m.PushAnchor("")
		}
	})
}
func _dream_show(m *ice.Message, name string) {
	if !strings.Contains(name, "-") || !strings.HasPrefix(name, "20") {
		name = m.Time("20060102-") + kit.ReplaceAll(name, "-", "_")
	}
	defer m.ProcessOpen(m.MergePOD(m.Option(mdb.NAME, name)))
	defer m.Echo(m.MergePOD(m.Option(mdb.NAME, name)))

	// 任务目录
	p := path.Join(m.Config(nfs.PATH), name)
	if m.Option(nfs.REPOS) != "" { // 下载源码
		m.Cmd("web.code.git.repos", mdb.CREATE, m.OptionSimple(nfs.REPOS), nfs.PATH, p)
	} else { // 创建目录
		os.MkdirAll(p, ice.MOD_DIR)
	}

	// 任务模板
	if m.Option(nfs.TEMPLATE) != "" {
		for _, file := range []string{ice.ETC_MISS_SH, ice.SRC_MAIN_SHY, ice.SRC_MAIN_GO, ice.GO_MOD, ice.MAKEFILE} {
			if _, e := os.Stat(path.Join(p, file)); os.IsNotExist(e) {
				switch m.Cmdy(nfs.COPY, path.Join(p, file), path.Join(m.Option(nfs.TEMPLATE), file)); file {
				case ice.GO_MOD:
					kit.Rewrite(path.Join(p, file), func(line string) string {
						return kit.Select(line, "module "+name, strings.HasPrefix(line, "module"))
					})
				}
			}
		}
	}

	// 任务脚本
	miss := path.Join(p, ice.ETC_MISS_SH)
	if _, e := os.Stat(miss); os.IsNotExist(e) {
		m.Cmd(nfs.SAVE, miss, m.Config("miss"))
	}
	defer m.Cmdy(nfs.DIR, p)

	if b, e := ioutil.ReadFile(path.Join(p, m.Conf(gdb.SIGNAL, kit.Keym(cli.PID)))); e == nil {
		if s, e := os.Stat("/proc/" + string(b)); e == nil && s.IsDir() {
			m.Info("already exists %v", string(b))
			return // 已经启动
		}
	}
	if m.Cmd(SPACE, name).Length() > 0 {
		return // 已经启动
	}

	m.ToastProcess()
	defer m.ToastSuccess()

	m.Optionv(cli.CMD_DIR, p)
	m.Optionv(cli.CMD_ENV, kit.Simple(
		cli.CTX_DEV, "http://:"+m.Cmd(SERVE, ice.OptionFields("")).Append(tcp.PORT),
		cli.PATH, kit.Path(path.Join(p, ice.BIN))+":"+kit.Path(ice.BIN)+":"+os.Getenv(cli.PATH),
		cli.USER, ice.Info.UserName, cli.HOME, os.Getenv(cli.HOME), m.Configv(cli.ENV),
	))
	m.Optionv(cli.CMD_OUTPUT, path.Join(p, m.Config(kit.Keys(cli.ENV, cli.CTX_LOG))))

	// 启动任务
	m.Cmd(cli.DAEMON, kit.Path(os.Args[0]), SPACE, tcp.DIAL, ice.DEV, ice.OPS, mdb.NAME, name, m.OptionSimple(RIVER))
	defer m.Event(DREAM_CREATE, kit.SimpleKV("", m.Option(mdb.TYPE), name)...)
	m.Sleep3s()
}

const (
	DREAM_CREATE = "dream.create"
	DREAM_START  = "dream.start"
	DREAM_STOP   = "dream.stop"
)
const DREAM = "dream"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		DREAM: {Name: "dream name path auto start", Help: "梦想家", Action: map[string]*ice.Action{
			cli.START: {Name: "start name repos river", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				_dream_show(m, m.Option(mdb.NAME, kit.Select(path.Base(m.Option(nfs.REPOS)), m.Option(mdb.NAME))))
			}},
			DREAM_STOP: {Name: "dream.stop type name", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd(SPACE, m.Option(mdb.NAME)).Append(mdb.STATUS) == cli.STOP {
					m.Cmdy(SPACE, mdb.REMOVE, m.OptionSimple(mdb.NAME))
					return
				}
				m.Cmdy(SPACE, mdb.REMOVE, m.OptionSimple(mdb.NAME))
				m.Sleep("1s", DREAM, cli.START, m.OptionSimple(mdb.NAME))
			}},
			cli.STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(SPACE, mdb.MODIFY, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				m.Cmdy(SPACE, m.Option(mdb.NAME), "exit", "0")
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				_dream_list(m).Cut("name,status,time")
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				_dream_list(m)
				return
			}

			m.Option(nfs.DIR_ROOT, path.Join(m.Config(nfs.PATH), arg[0]))
			m.Cmdy(nfs.CAT, arg[1:])
		}},
	}, Configs: map[string]*ice.Config{
		DREAM: {Name: DREAM, Help: "梦想家", Value: kit.Data(nfs.PATH, ice.USR_LOCAL_WORK,
			cli.ENV, kit.Dict(cli.CTX_LOG, ice.BIN_BOOT_LOG),
			"miss", `#!/bin/bash
if [ "$ISH_CONF_PRE" = "" ]; then
	[ -f $PWD/.ish/plug.sh ] || [ -f $HOME/.ish/plug.sh ] || git clone ${ISH_CONF_HUB_PROXY:="https://"}shylinux.com/x/intshell $PWD/.ish
	source $PWD/.ish/plug.sh || source $HOME/.ish/plug.sh
fi

require miss.sh
ish_miss_prepare_compile
ish_miss_prepare_develop
ish_miss_prepare_install

# ish_miss_prepare wubi-dict
# ish_miss_prepare word-dict

# ish_miss_prepare linux-story
# ish_miss_prepare mysql-story
# ish_miss_prepare release

ish_miss_prepare_contexts
ish_miss_prepare_intshell
# ish_miss_prepare_icebergs
# ish_miss_prepare_toolkits
# ish_miss_prepare_volcanos
# ish_miss_prepare_learning

make
`,
		)},
	}})
}
