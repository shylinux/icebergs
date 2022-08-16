package bash

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

func _run_action(m *ice.Message, cmd *ice.Command, script string, arg ...string) {
	m.SetResult().Echo("#/bin/bash\n")

	list, args := []string{}, []string{}
	kit.Fetch(cmd.Meta["_trans"], func(k string, v string) {
		list = append(list, k)
		args = append(args, kit.Format(`			%s)`, k))
		kit.Fetch(cmd.Meta[k], func(index int, value ice.Map) {
			args = append(args, kit.Format(`				read -p "input %s: " v; url="$url/%s/$v" `, value[mdb.NAME], value[mdb.NAME]))
		})
		args = append(args, kit.Format(`				;;`))
	})

	m.Echo(`
ish_sys_dev_run_source() {
	select action in %s; do
		local url="run/action/run/%s/action/$action"
		case $action in
%s
		esac
		ish_sys_dev_source $url
	done
}
`, kit.Join(list, ice.SP), arg[0], kit.Join(args, ice.NL))

	m.Echo(`
ish_sys_dev_run_action() {
	select action in %s; do
		local url="run/action/run/%s/action/$action"
		case $action in
%s
		esac
		ish_sys_dev_request $url
		echo
	done
}
`, kit.Join(list, ice.SP), arg[0], kit.Join(args, ice.NL))

	m.Echo(`
ish_sys_dev_run_command() {
	ish_sys_dev_run %s "$@"
}
`, arg[0])

	m.Echo(kit.Select("cat $1", script))
}

const RUN = "run"

func init() {
	Index.MergeCommands(ice.Commands{
		"/run/": {Name: "/run/", Help: "执行", Actions: ice.Actions{
			ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Search(arg[0], func(_ *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
					if p := strings.ReplaceAll(kit.Select("/app/cat.sh", cmd.Meta[ctx.DISPLAY]), ".js", ".sh"); strings.HasPrefix(p, ice.PS+ice.REQUIRE) {
						m.Cmdy(web.SPIDE, ice.DEV, web.SPIDE_RAW, p)
					} else {
						m.Cmdy(nfs.CAT, path.Join(ice.USR_INTSHELL, p))
					}
					if m.IsErrNotFound() {
						m.SetResult()
					}
					_run_action(m, cmd, m.Result(), arg...)
				})
			}},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				if !ctx.PodCmd(m, arg) && aaa.Right(m, arg) {
					m.Cmdy(arg)
				}
				if m.Result() == "" {
					m.Table()
				}
			}},
		}},
	})
}
