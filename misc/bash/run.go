package bash

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const RUN = "run"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		"/run/": {Name: "/run/", Help: "执行", Action: ice.MergeAction(map[string]*ice.Action{
			ctx.COMMAND: {Name: "command", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				m.Search(arg[0], func(_ *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
					p := strings.ReplaceAll(kit.Select("/app/cat.sh", cmd.Meta["display"]), ".js", ".sh")
					if strings.HasPrefix(p, ice.PS+ice.REQUIRE) {
						m.Cmdy(web.SPIDE, ice.DEV, web.SPIDE_RAW, p)
					} else {
						m.Cmdy(nfs.CAT, path.Join(ice.USR_INTSHELL, p))
					}
					m.Debug(kit.Formats(cmd.Meta))
					if m.Result() == "" || m.Result(1) == ice.ErrNotFound {
						m.Set(ice.MSG_RESULT)
						m.Echo("#/bin/bash\n")
						list := []string{}
						args := []string{}
						kit.Fetch(cmd.Meta["_trans"], func(k string, v string) {
							list = append(list, k)
							args = append(args, kit.Format(`			%s)`, k))
							kit.Fetch(cmd.Meta[k], func(index int, value map[string]interface{}) {
								args = append(args, kit.Format(`				read -p "read %s: " v; url="$url/%s/$v" `, value[kit.MDB_NAME], value[kit.MDB_NAME]))
							})
							args = append(args, kit.Format(`				;;`))
						})
						list = append(list, "quit")
						m.Echo(`
ish_sys_dev_run_action() {
	select action in %s; do
		if [ "$action" = "quit" ]; then break; fi
		local url="run/action/run/%s/action/$action"
		case $action in
%s
		esac
		ish_sys_dev_request $url
		echo
	done
}
`, kit.Join(list, ice.SP), arg[0], kit.Join(args, ice.NL))
						m.Echo("cat $1\n")
						m.Debug("what %v", m.Result())
					}
				})
			}},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				if m.Right(arg) && !m.PodCmd(arg) {
					m.Cmdy(arg)
				}
				if m.Result() == "" {
					m.Table()
				}
			}},
		}, ctx.CmdAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if m.Right(arg) {
				m.Cmdy(arg)
			}
		}}},
	})
}
