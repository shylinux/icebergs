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

func _sh_exec(m *ice.Message, arg ...string) (res []string) {
	if cmd := ctx.GetFileCmd(path.Join(arg[2], arg[1])); cmd != "" {
		res = append(res, kit.Format(_sh_template, "http://localhost:9020", m.Option(ice.MSG_USERPOD), cmd))
	}
	if _, e := nfs.DiskFile.StatFile(path.Join(arg[2], arg[1])); e == nil {
		res = append(res, kit.Format("source %s", kit.Path(arg[2], arg[1])))
	} else if b, e := nfs.ReadFile(m, path.Join(arg[2], arg[1])); e == nil {
		res = append(res, string(b))
	}
	m.Cmdy(cli.SYSTEM, SH, "-c", kit.Join(res, ice.NL)).StatusTime()
	return
}

const SH = nfs.SH

func init() {
	Index.MergeCommands(ice.Commands{
		SH: {Name: "sh path auto", Help: "命令", Actions: ice.MergeActions(ice.Actions{
			mdb.RENDER: {Hand: func(m *ice.Message, arg ...string) { _c_show(m, arg...) }},
			mdb.ENGINE: {Hand: func(m *ice.Message, arg ...string) { _sh_exec(m, arg...) }},
			NAVIGATE:   {Hand: func(m *ice.Message, arg ...string) { _c_tags(m, MAN, "ctags", "-a", "-R", nfs.PWD) }},
		}, PlugAction())},
	})
}

var _sh_template = `#! /bin/sh
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
`
