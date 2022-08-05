package code

import (
	"path"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _sh_main_script(m *ice.Message, arg ...string) (res []string) {
	if cmd := ctx.GetFileCmd(path.Join(arg[2], arg[1])); cmd != "" {
		res = append(res, kit.Format(`#! /bin/sh
export ctx_dev=%s; ctx_pod=%s ctx_temp=$(mktemp); curl -fsSL $ctx_dev -o $ctx_temp; source $ctx_temp &>/dev/null
_done=""
_list() {
	if [ "$_done" = "" ]; then
		ish_sys_dev_run %s "$@"
	else
		ish_sys_dev_run_command "$@"
	fi
	_done=done
}
_action() {
	_list action "$@"
}
`, "http://localhost:9020", m.Option(ice.MSG_USERPOD), cmd))
	}

	if _, e := nfs.DiskFile.StatFile(path.Join(arg[2], arg[1])); e == nil {
		res = append(res, kit.Format("source %s", kit.Path(arg[2], arg[1])))
	} else if b, e := nfs.ReadFile(m, path.Join(arg[2], arg[1])); e == nil {
		res = append(res, string(b))
	}
	m.Cmdy(cli.SYSTEM, SH, "-c", kit.Join(res, ice.NL))
	if m.StatusTime(); cli.IsSuccess(m) {
		m.SetAppend()
	}
	return
}

func _sh_exec(m *ice.Message, arg ...string) {
	if m.Option(mdb.TEXT) == "" {
		// if _cache_bin != nil {
		// 	m.Copy(_cache_bin)
		// 	break
		// }
		// _cache_bin = m

		// m.Push(mdb.NAME, "_list")
		// _vimer_list(m, "/bin")
		// _vimer_list(m, "/sbin")
	}
}

const SH = nfs.SH

func init() {
	Index.Register(&ice.Context{Name: SH, Help: "命令", Commands: ice.Commands{
		SH: {Name: "sh path auto", Help: "命令", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Name: "_init", Help: "初始化", Hand: func(m *ice.Message, arg ...string) {
				for _, cmd := range []string{mdb.SEARCH, mdb.ENGINE, mdb.RENDER, mdb.PLUGIN} {
					m.Cmd(cmd, mdb.CREATE, m.CommandKey(), m.PrefixKey())
				}
				LoadPlug(m, m.CommandKey())
			}},
			mdb.SEARCH: {Name: "search", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == mdb.FOREACH {
					return
				}
				m.Option(cli.CMD_DIR, kit.Select(ice.SRC, arg, 2))
				m.Cmdy(mdb.SEARCH, MAN1, arg[1:])
				m.Cmdy(mdb.SEARCH, MAN8, arg[1:])
				_go_find(m, kit.Select(cli.MAIN, arg, 1), arg[2])
				_go_grep(m, kit.Select(cli.MAIN, arg, 1), arg[2])
			}},
			mdb.ENGINE: {Name: "engine", Help: "引擎", Hand: func(m *ice.Message, arg ...string) {
				_sh_exec(m, arg...)
			}},
			mdb.RENDER: {Name: "render", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
				_sh_main_script(m, arg...)
			}},
		}, PlugAction()), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 0 && kit.Ext(arg[0]) == SH {
				_sh_main_script(m, SH, arg[0], ice.SRC)
				return
			}
			m.Option(nfs.DIR_ROOT, ice.SRC)
			m.Option(nfs.DIR_DEEP, ice.TRUE)
			m.Option(nfs.DIR_REG, ".*.(sh)$")
			m.Cmdy(nfs.DIR, arg)
		}},
	}, Configs: ice.Configs{
		SH: {Name: SH, Help: "命令", Value: kit.Data(PLUG, kit.Dict(
			SPLIT, kit.Dict(SPACE, " ", OPERATE, "{[(.,;!|<>)]}"),
			PREFIX, kit.Dict("#!", COMMENT, "# ", COMMENT), SUFFIX, kit.Dict(" {", COMMENT),
			PREPARE, kit.Dict(
				KEYWORD, kit.Simple(
					"require", "source", "return", "local", "export", "env",

					"if", "then", "else", "fi",
					"for", "while", "do", "done",
					"esac", "case", "in",

					"shift",
					"echo",
					"read",
					"eval",
					"kill",
					"let",
					"cd",
				),
				FUNCTION, kit.Simple(
					"xargs", "_list",
					"date", "uptime", "uname", "whoami",
					"find", "grep", "sed", "awk",
					"pwd",
					"ls",
					"ps",
					"rm",
					"go",
				),
			), KEYWORD, kit.Dict(),
		))},
	}}, nil)
}
