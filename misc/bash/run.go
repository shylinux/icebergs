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

const RUN = "run"

func init() {
	Index.MergeCommands(ice.Commands{
		web.PP(RUN): {Actions: ice.Actions{
			"check": {Name: "check sid*", Hand: func(m *ice.Message, arg ...string) { m.Echo(m.Cmd(SESS, m.Option(SID)).Append(GRANT)) }},
			"complete": {Hand: func(m *ice.Message, arg ...string) {
				list := kit.Split(m.Option("line"))[1:]
				if len(list) == kit.Int(m.Option("cword")) {
					list = kit.Slice(list, 0, -1)
				}
				m.Echo(strings.Join(Complete(m, false, list...), ice.NL))
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

func Complete(m *ice.Message, detail bool, arg ...string) (res []string) {
	echo := func(arg ...string) { res = append(res, arg...) }
	if len(arg) < 2 || arg[1] != ctx.ACTION {
		list := ctx.CmdList(m.Spawn()).Appendv(ctx.INDEX)
		if len(arg) > 0 {
			pre := arg[0][0 : strings.LastIndex(arg[0], ice.PT)+1]
			list = kit.Simple(list, func(cmd string) bool { return strings.HasPrefix(cmd, arg[0]) }, func(cmd string) string { return strings.TrimPrefix(cmd, pre) })
		}
		if len(arg) > 1 || (len(list) == 1 && kit.Select("", kit.Split(arg[0], ice.PT), -1) == list[0]) {
			kit.If(detail, func() { echo("func") })
			m.Cmdy(arg).Search(arg[0], func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
				field := kit.Format(kit.Value(cmd.List, kit.Keys(len(arg)-1, mdb.NAME)))
				m.Table(func(index int, value ice.Maps, head []string) {
					echo(value[field])
					if detail {
						echo(kit.Join(kit.Simple(head, func(key string) string { return key + ": " + value[key] }), ice.SP))
					}
				})
			})
			kit.If(len(arg) == 1, func() { echo(ctx.ACTION) })
		} else {
			echo(list...)
		}
	} else {
		m.Search(arg[0], func(p *ice.Context, s *ice.Context, key string, cmd *ice.Command) {
			if len(arg) > 2 && cmd.Actions != nil {
				if _, ok := cmd.Actions[arg[2]]; ok {
					if len(arg)%2 == 1 {
						list := map[string]bool{}
						kit.For(arg[3:], func(k, v string) { list[k] = true })
						kit.For(cmd.Meta[arg[2]], func(value ice.Map) {
							if field := kit.Format(value[mdb.NAME]); !list[field] {
								echo(field)
							}
						})
					} else {
						m.Options(arg[3:])
						m.Cmdy(arg[0], mdb.INPUTS, kit.Select("", arg, -1)).Tables(func(value ice.Maps) {
							v := value[m.Appendv(ice.MSG_APPEND)[0]]
							kit.If(strings.Contains(v, ice.SP), func() { echo("\"" + v + "\"") }, func() { echo(v) })
						})
					}
					return
				}
			}
			if len(arg) < 4 {
				kit.If(detail, func() { echo("func") })
				kit.For(kit.SortedKey(cmd.Actions), func(sub string) {
					if strings.HasPrefix(sub, kit.Select("", arg, 2)) {
						if echo(sub); detail {
							echo(cmd.Actions[sub].Name + ice.SP + cmd.Actions[sub].Help)
						}
					}
				})
			}
		})
	}
	return
}
