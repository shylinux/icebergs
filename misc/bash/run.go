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
	list, args := []string{}, []string{}
	kit.Fetch(cmd.Meta["_trans"], func(k string, v string) {
		list = append(list, k)
		args = append(args, kit.Format(`			%s)`, k))
		kit.Fetch(cmd.Meta[k], func(index int, value ice.Map) {
			args = append(args, kit.Format(`				read -p "input %s: " v; url="$url/%s/$v" `, value[mdb.NAME], value[mdb.NAME]))
		})
		args = append(args, kit.Format(`				;;`))
	})
	m.SetResult().Echo("#/bin/sh\n")
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

func _run_command(m *ice.Message, key string, arg ...string) {
	m.Search(key, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		m.Echo(kit.Join(m.Cmd(key, arg).Appendv(kit.Format(kit.Value(cmd.List, kit.Keys(len(arg), mdb.NAME)))), ice.SP))
	})
}
func _run_actions(m *ice.Message, key, sub string, arg ...string) (res []string) {
	m.Search(key, func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
		if sub == "" {
			res = kit.SortedKey(cmd.Meta)
		} else if len(arg)%2 == 0 {
			kit.Fetch(kit.Value(cmd.Meta, sub), func(value ice.Map) { res = append(res, kit.Format(value[mdb.NAME])) })
			kit.Fetch(arg, func(k, v string) { res[kit.IndexOf(res, k)] = "" })
		} else {
			msg := m.Cmd(key, mdb.INPUTS, kit.Select("", arg, -1), kit.Dict(arg))
			res = msg.Appendv(kit.Select("", msg.Appendv(ice.MSG_APPEND), 0))
		}
		m.Echo(kit.Join(res, ice.SP))
	})
	return nil
}

const RUN = "run"

func init() {
	Index.MergeCommands(ice.Commands{
		web.PP(RUN): {Actions: ice.Actions{
			"check": {Name: "check sid*", Hand: func(m *ice.Message, arg ...string) { m.Echo(m.Cmd(SESS, m.Option(SID)).Append(GRANT)) }},
			"complete": {Hand: func(m *ice.Message, arg ...string) {
				switch arg = kit.Split(m.Option("line")); m.Option("cword") {
				case "1":
					m.Echo("action ")
					m.Echo(strings.Join(m.Cmd(ctx.COMMAND, mdb.SEARCH, ctx.COMMAND, ice.OptionFields(ctx.INDEX)).Appendv(ctx.INDEX), ice.SP))
				default:
					if kit.Int(m.Option("cword"))+1 == len(arg) {
						arg = kit.Slice(arg, 0, -1)
					}
					if kit.Select("", arg, 2) == ctx.ACTION {
						_run_actions(m, arg[1], kit.Select("", arg, 3), kit.Slice(arg, 4)...)
					} else {
						_run_command(m, arg[1], arg[2:]...)
					}
				}
			}},
			ctx.COMMAND: {Hand: func(m *ice.Message, arg ...string) {
				m.Search(arg[0], func(_ *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
					if p := kit.ExtChange(kit.Select("/app/cat.sh", cmd.Meta[ctx.DISPLAY]), nfs.SH); strings.HasPrefix(p, ice.PS+ice.REQUIRE) {
						m.Cmdy(web.SPIDE, ice.DEV, web.SPIDE_RAW, p)
					} else {
						m.Cmdy(nfs.CAT, path.Join(ice.USR_INTSHELL, p))
					}
					_run_action(m, cmd, m.Results(), arg...)
				})
			}},
			ice.RUN: {Hand: func(m *ice.Message, arg ...string) {
				if !ctx.PodCmd(m, arg) && aaa.Right(m, arg) {
					m.Cmdy(arg)
				}
				if m.Result() != "" && !strings.HasSuffix(m.Result(), ice.NL) {
					m.Echo(ice.NL)
				}
			}},
		}},
	})
}
