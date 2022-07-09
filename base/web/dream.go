package web

import (
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
	return m.Cmdy(nfs.DIR, m.Config(nfs.PATH), "time,size,name").Table(func(index int, value ice.Maps, head []string) {
		if m.Richs(SPACE, nil, value[mdb.NAME], func(key string, val ice.Map) {
			m.Push(mdb.TYPE, val[mdb.TYPE])
			m.Push(cli.STATUS, cli.START)
			m.PushButton("vimer", cli.OPEN, cli.STOP)
			m.PushAnchor(strings.Split(m.MergePod(value[mdb.NAME]), "?")[0])
		}) == nil {
			m.Push(mdb.TYPE, WORKER)
			m.Push(cli.STATUS, cli.STOP)
			m.PushButton(cli.START, nfs.TRASH)
			m.PushAnchor("")
		}
	})
}
func _dream_show(m *ice.Message, name string) {
	if !strings.Contains(name, "-") || !strings.HasPrefix(name, "20") {
		name = m.Time("20060102-") + kit.ReplaceAll(name, "-", "_")
	}
	defer m.ProcessOpen(m.MergePod(m.Option(mdb.NAME, name)))
	defer m.Echo(m.MergePod(m.Option(mdb.NAME, name)))

	// 任务目录
	p := path.Join(m.Config(nfs.PATH), name)
	if m.Option(nfs.REPOS) != "" { // 下载源码
		m.Cmd("web.code.git.repos", mdb.CREATE, m.OptionSimple(nfs.REPOS), nfs.PATH, p)
	} else { // 创建目录
		nfs.MkdirAll(m, p)
	}

	// 任务模板
	if m.Option(nfs.TEMPLATE) != "" {
		for _, file := range []string{
			ice.ETC_MISS_SH, ice.SRC_MAIN_SHY, ice.SRC_MAIN_GO,
			ice.GO_MOD, ice.MAKEFILE, ice.README_MD,
		} {
			if kit.FileExists(path.Join(p, file)) {
				continue
			}
			switch m.Cmdy(nfs.COPY, path.Join(p, file), path.Join(m.Config(nfs.PATH), m.Option(nfs.TEMPLATE), file)); file {
			case ice.GO_MOD:
				kit.Rewrite(path.Join(p, file), func(line string) string {
					return kit.Select(line, "module "+name, strings.HasPrefix(line, "module"))
				})
			}
		}
	}

	// 任务脚本
	m.Cmd(nfs.DEFS, path.Join(p, ice.ETC_MISS_SH), m.Config("miss"))
	defer m.Cmdy(nfs.DIR, p)

	if pid := m.Cmdx(nfs.CAT, path.Join(p, m.Conf(gdb.SIGNAL, kit.Keym(nfs.PATH)))); pid != "" && kit.FileExists("/proc/"+pid) {
		m.Info("already exists %v", pid)
		return // 已经启动
	} else if m.Cmd(SPACE, name).Length() > 0 {
		return // 已经启动
	}

	defer m.ToastProcess()()

	m.Optionv(cli.CMD_DIR, p)
	m.Optionv(cli.CMD_ENV, kit.Simple(
		cli.CTX_OPS, "http://:"+m.Cmd(SERVE, ice.OptionFields("")).Append(tcp.PORT),
		cli.PATH, cli.BinPath(kit.Path(p, ice.BIN)), cli.HOME, kit.Env(cli.HOME),
		cli.SHELL, kit.Env(cli.SHELL), cli.TERM, kit.Env(cli.TERM),
		cli.USER, ice.Info.UserName, m.Configv(cli.ENV),
	))
	m.Optionv(cli.CMD_OUTPUT, path.Join(p, ice.BIN_BOOT_LOG))

	// 启动任务
	bin := kit.Select(os.Args[0], cli.SystemFind(m, ice.ICE_BIN, kit.Path(path.Join(p, ice.BIN)), kit.Path(ice.BIN)))
	m.Cmd(cli.DAEMON, bin, SPACE, tcp.DIAL, ice.DEV, ice.OPS, m.OptionSimple(mdb.NAME, RIVER))

	m.Sleep3s()
	m.Option(cli.CMD_ENV, "")
	m.Option(cli.CMD_OUTPUT, "")
	m.Event(DREAM_CREATE, kit.SimpleKV("", m.Option(mdb.TYPE), name)...)
}

const (
	DREAM_CREATE = "dream.create"
	DREAM_START  = "dream.start"
	DREAM_STOP   = "dream.stop"
)
const DREAM = "dream"

func init() {
	Index.Merge(&ice.Context{Commands: ice.Commands{
		DREAM: {Name: "dream name path auto start", Help: "梦想家", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Config("miss", _dream_miss)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case "repos":
				default:
					_dream_list(m).Cut("name,status,time")
				}
			}},
			cli.START: {Name: "start name=hi repos", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				_dream_show(m, m.Option(mdb.NAME, kit.Select(path.Base(m.Option(nfs.REPOS)), m.Option(mdb.NAME))))
			}},
			cli.OPEN: {Name: "open", Help: "打开", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(m.MergePod(m.Option(mdb.NAME), "", ""))
			}},
			"vimer": {Name: "vimer", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(m.MergePod(m.Option(mdb.NAME)+"/cmd/web.code.vimer", "", ""))
			}},
			cli.STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPACE, mdb.MODIFY, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				m.Cmd(SPACE, m.Option(mdb.NAME), ice.EXIT)
				m.ProcessRefresh("100ms")
			}},
			DREAM_STOP: {Name: "dream.stop type name", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				if m.Cmd(SPACE, m.Option(mdb.NAME)).Append(mdb.STATUS) == cli.STOP {
					m.Cmd(mdb.DELETE, m.Prefix(SPACE), "", mdb.HASH, m.OptionSimple(mdb.NAME))
				} else {
					m.Cmd(mdb.DELETE, m.Prefix(SPACE), "", mdb.HASH, m.OptionSimple(mdb.NAME))
					m.Sleep("1s", DREAM, cli.START, m.OptionSimple(mdb.NAME))
				}
			}},
			nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.TRASH, mdb.CREATE, path.Join(m.Config(nfs.PATH), m.Option(mdb.NAME)))
				m.ProcessRefresh30ms()
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			if start := 0; len(arg) == 0 {
				_dream_list(m).SetAppend(mdb.TEXT)
				m.Table(func(index int, value ice.Maps, head []string) {
					if value[cli.STATUS] != cli.START {
						m.Push(mdb.TEXT, "")
						return
					}
					start++
					text := []string{}
					for _, line := range kit.Split(m.Cmdx(SPACE, value[mdb.NAME], cli.SYSTEM, "git", "diff", "--shortstat"), ice.FS, ice.FS) {
						if list := kit.Split(line); strings.Contains(line, "file") {
							text = append(text, list[0]+" file")
						} else if strings.Contains(line, "ins") {
							text = append(text, list[0]+" +++")
						} else if strings.Contains(line, "dele") {
							text = append(text, list[0]+" ---")
						}
					}
					m.Push(mdb.TEXT, strings.Join(text, ", "))
				}).Sort("status,type,name").StatusTimeCount(cli.START, start)
				if !m.IsMobileUA() {
					m.Display("/plugin/table.js?style=card")
				}
				return
			}

			m.Option(nfs.DIR_ROOT, path.Join(m.Config(nfs.PATH), arg[0]))
			m.Cmdy(nfs.CAT, arg[1:])
		}},
	}, Configs: ice.Configs{
		DREAM: {Name: DREAM, Help: "梦想家", Value: kit.Data(nfs.PATH, ice.USR_LOCAL_WORK, "miss", _dream_miss)},
	}})
}

var _dream_miss = `#! /bin/sh

require &>/dev/null || if [ -f $PWD/.ish/plug.sh ]; then source $PWD/.ish/plug.sh; elif [ -f $HOME/.ish/plug.sh ]; then source $HOME/.ish/plug.sh; else
	ctx_temp=$(mktemp); if curl -h &>/dev/null; then curl -o $ctx_temp -fsSL https://shylinux.com; else wget -O $ctx_temp -q http://shylinux.com; fi; source $ctx_temp intshell
fi

require miss.sh
ish_miss_prepare_compile
ish_miss_prepare_develop
ish_miss_prepare_install

ish_miss_make; if [ -n "$*" ]; then ./bin/ice.bin forever serve "$@"; fi
`
