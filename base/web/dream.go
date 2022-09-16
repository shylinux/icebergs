package web

import (
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/gdb"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	kit "shylinux.com/x/toolkits"
)

func _dream_list(m *ice.Message) *ice.Message {
	list := m.CmdMap(SPACE, mdb.NAME)
	m.Cmdy(nfs.DIR, ice.USR_LOCAL_WORK+ice.PS, "time,size,name", func(value ice.Maps) {
		if dream, ok := list[value[mdb.NAME]]; ok {
			m.Push(mdb.TYPE, dream[mdb.TYPE])
			m.Push(cli.STATUS, cli.START)
			m.PushButton(cli.OPEN, "vimer", "xterm", cli.STOP)
			m.PushAnchor(strings.Split(MergePod(m, value[mdb.NAME]), "?")[0])
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
		} else {
			m.Push(mdb.TYPE, WORKER)
			m.Push(cli.STATUS, cli.STOP)
			m.PushButton(cli.START, nfs.TRASH)
			m.PushAnchor("")
			m.Push(mdb.TEXT, "")
		}
	})
	m.Sort("status,type,name")
	m.StatusTimeCount(cli.START, len(list))
	return m
}

func _dream_show(m *ice.Message, name string) {
	if m.Warn(name == "") {
		return
	}
	if !strings.Contains(name, "-") || !strings.HasPrefix(name, "20") {
		name = m.Time("20060102-") + name
	}
	defer m.ProcessOpen(MergePod(m, m.Option(mdb.NAME, name)))

	p := path.Join(ice.USR_LOCAL_WORK, name)
	if pid := m.Cmdx(nfs.CAT, path.Join(p, ice.Info.PidPath)); pid != "" && nfs.ExistsFile(m, "/proc/"+pid) {
		m.Info("already exists %v", pid)
		return // 已经启动
	} else if m.Cmd(SPACE, name).Length() > 0 {
		return // 已经启动
	}

	if m.Option(nfs.REPOS) != "" { // 下载源码
		m.Cmd("web.code.git.repos", mdb.CREATE, m.OptionSimple(nfs.REPOS), nfs.PATH, p)
	} else { // 创建目录
		nfs.MkdirAll(m, p)
	}

	// 目录文件
	if m.Option(nfs.TEMPLATE) != "" {
		for _, file := range []string{
			ice.ETC_MISS_SH, ice.SRC_MAIN_SHY, ice.SRC_MAIN_GO,
			ice.GO_MOD, ice.MAKEFILE, ice.README_MD,
		} {
			if nfs.ExistsFile(m, path.Join(p, file)) {
				continue
			}
			switch m.Cmdy(nfs.COPY, path.Join(p, file), path.Join(ice.USR_LOCAL_WORK, m.Option(nfs.TEMPLATE), file)); file {
			case ice.GO_MOD:
				kit.Rewrite(path.Join(p, file), func(line string) string {
					return kit.Select(line, "module "+name, strings.HasPrefix(line, "module"))
				})
			}
		}
	}
	m.Cmd(nfs.DEFS, path.Join(p, ice.ETC_MISS_SH), m.Config(nfs.SCRIPT))

	// 环境变量
	m.Optionv(cli.CMD_DIR, kit.Path(p))
	m.Optionv(cli.CMD_ENV, kit.Simple(
		cli.CTX_OPS, "http://:"+m.Cmd(SERVE, ice.OptionFields("")).Append(tcp.PORT),
		cli.PATH, cli.BinPath(kit.Path(p, ice.BIN)), cli.HOME, kit.Env(cli.HOME),
		cli.TERM, kit.Env(cli.TERM), cli.SHELL, kit.Env(cli.SHELL),
		cli.USER, ice.Info.UserName, m.Configv(cli.ENV),
	))
	m.Optionv(cli.CMD_OUTPUT, path.Join(p, ice.BIN_BOOT_LOG))

	defer ToastProcess(m)()
	bin := kit.Select(os.Args[0], cli.SystemFind(m, ice.ICE_BIN, nfs.PWD+path.Join(p, ice.BIN), nfs.PWD+ice.BIN))
	m.Cmd(cli.DAEMON, bin, SPACE, tcp.DIAL, ice.DEV, ice.OPS, m.OptionSimple(mdb.NAME, RIVER), "daemon", "ops")

	defer gdb.Event(m, DREAM_CREATE, m.OptionSimple(mdb.TYPE, mdb.NAME))
	m.Option(cli.CMD_OUTPUT, "")
	m.Option(cli.CMD_ENV, "")
	m.Sleep3s()
}

const (
	DREAM_CREATE = "dream.create"
	DREAM_START  = "dream.start"
	DREAM_STOP   = "dream.stop"
)
const DREAM = "dream"

func init() {
	Index.MergeCommands(ice.Commands{
		DREAM: {Name: "dream name path auto create", Help: "梦想家", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { m.Config(nfs.SCRIPT, _dream_script) }},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				switch arg[0] {
				case nfs.REPOS:
					m.Cmd("web.code.git.server", func(value ice.Maps) {
						m.Push(nfs.PATH, MergeLink(m, path.Join("/x/", path.Clean(value[nfs.PATH])+".git")))
					})
					m.Sort(nfs.PATH)
				default:
					_dream_list(m).Cut("name,status,time")
				}
			}},
			mdb.CREATE: {Name: "create name=hi repos", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				_dream_show(m, m.Option(mdb.NAME, kit.Select(path.Base(m.Option(nfs.REPOS)), m.Option(mdb.NAME))))
			}},
			cli.START: {Name: "start", Help: "启动", Hand: func(m *ice.Message, arg ...string) {
				_dream_show(m, m.Option(mdb.NAME))
			}},
			cli.OPEN: {Name: "open", Help: "打开", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(MergePod(m, m.Option(mdb.NAME), "", ""))
			}},
			"vimer": {Name: "vimer", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(MergePod(m, m.Option(mdb.NAME)+"/cmd/web.code.vimer", "", ""))
			}},
			"xterm": {Name: "xterm", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.ProcessOpen(MergePod(m, m.Option(mdb.NAME)+"/cmd/web.code.xterm", mdb.HASH,
					m.Cmdx(SPACE, m.Option(mdb.NAME), "web.code.xterm", mdb.CREATE, mdb.TYPE, nfs.SH, mdb.NAME, "xterm")))
			}},
			cli.STOP: {Name: "stop", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(SPACE, mdb.MODIFY, m.OptionSimple(mdb.NAME), mdb.STATUS, cli.STOP)
				m.Go(func() { m.Cmd(SPACE, m.Option(mdb.NAME), ice.EXIT) })
				m.Sleep3s()
			}},
			DREAM_STOP: {Name: "dream.stop type name", Help: "停止", Hand: func(m *ice.Message, arg ...string) {
				if m.CmdAppend(SPACE, m.Option(mdb.NAME), mdb.STATUS) == cli.STOP {
					m.Cmd(mdb.DELETE, m.Prefix(SPACE), "", mdb.HASH, m.OptionSimple(mdb.NAME))
				} else {
					m.Cmd(mdb.DELETE, m.Prefix(SPACE), "", mdb.HASH, m.OptionSimple(mdb.NAME))
					m.Sleep3s(DREAM, cli.START, m.OptionSimple(mdb.NAME))
				}
			}},
			nfs.TRASH: {Name: "trash", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(nfs.TRASH, mdb.CREATE, path.Join(ice.USR_LOCAL_WORK, m.Option(mdb.NAME)))
				m.ProcessRefresh30ms()
			}},
		}, mdb.HashAction(nfs.SCRIPT, _dream_script)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				if _dream_list(m); !m.IsMobileUA() {
					ctx.DisplayTableCard(m)
				}
			} else {
				m.Cmdy(nfs.CAT, arg[1:], kit.Dict(nfs.DIR_ROOT, path.Join(ice.USR_LOCAL_WORK, arg[0])))
			}
		}},
	})
}

var _dream_script = `#! /bin/sh

if [ -f $PWD/.ish/plug.sh ]; then source $PWD/.ish/plug.sh; elif [ -f $HOME/.ish/plug.sh ]; then source $HOME/.ish/plug.sh; else
	ctx_temp=$(mktemp); if curl -h &>/dev/null; then curl -o $ctx_temp -fsSL https://shylinux.com; else wget -O $ctx_temp -q http://shylinux.com; fi; source $ctx_temp intshell
fi

require miss.sh
ish_miss_prepare_compile
ish_miss_prepare_develop
ish_miss_prepare_operate

ish_miss_make; if [ -n "$*" ]; then ish_miss_serve "$@"; fi
`
