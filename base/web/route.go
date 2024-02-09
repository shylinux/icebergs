package web

import (
	"regexp"
	"runtime"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/tcp"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _route_push(m *ice.Message, space string, msg *ice.Message) *ice.Message {
	return msg.Table(func(index int, value ice.Maps, head []string) {
		value[SPACE], head = space, append(head, SPACE)
		m.Push("", value, head)
	})
}
func _route_match(m *ice.Message, space string, cb func(ice.Maps, int, []ice.Maps)) {
	reg, err := regexp.Compile(space)
	if m.WarnNotValid(err) {
		return
	}
	res := []ice.Maps{}
	list := kit.Split(space)
	m.Cmd("").Table(func(value ice.Maps) {
		if value[mdb.STATUS] == OFFLINE {

		} else if value[SPACE] == space {
			res = append(res, value)
		} else if kit.IsIn(value[SPACE], list...) {
			res = append(res, value)
		} else if reg.MatchString(kit.Format("%s:%s=%s@%s", value[SPACE], value[mdb.TYPE], value[nfs.MODULE], value[nfs.VERSION])) {
			res = append(res, value)
		}
	})
	for i, item := range res {
		cb(item, i, res)
	}
}
func _route_toast(m *ice.Message, space string, args ...string) {
	GoToast(m, "", func(toast func(string, int, int)) (list []string) {
		count, total := 0, 1
		_route_match(m, space, func(value ice.Maps, i int, _list []ice.Maps) {
			count, total = i, len(_list)
			toast(value[SPACE], count, total)
			if msg := _route_push(m, value[SPACE], m.Cmd(SPACE, value[SPACE], args, ice.Maps{ice.MSG_DAEMON: ""})); msg.IsErr() || !cli.IsSuccess(msg) {
				list = append(list, value[SPACE]+": "+msg.Result())
			} else {
				kit.If(msg.Result() == "", func() { msg.TableEcho() })
				m.Push(SPACE, value[SPACE]).Push(ice.RES, msg.Result())
			}
		})
		m.StatusTimeCount(ice.CMD, kit.Join(args, lex.SP), ice.SUCCESS, kit.Format("%d/%d", total-len(list), total))
		return
	})
}

const (
	ONLINE  = "online"
	OFFLINE = "offline"
)
const ROUTE = "route"

func init() {
	Index.MergeCommands(ice.Commands{
		ROUTE: {Name: "route space:text cmds:text auto build travel diagram prunes", Help: "路由表", Actions: ice.MergeActions(ice.Actions{
			ice.MAIN: {Help: "首页", Hand: func(m *ice.Message, arg ...string) { ctx.ProcessField(m, CHAT_IFRAME, m.MergePod(""), arg...) }},
			mdb.INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				switch mdb.HashInputs(m, arg); arg[0] {
				case SPACE:
					list := map[string]bool{}
					push := func(key string) { kit.If(!list[key], func() { m.Push(arg[0], key); list[key] = true }) }
					mdb.HashSelect(m.Spawn()).Table(func(value ice.Maps) { push(kit.Format("=%s@", value[nfs.MODULE])) })
					kit.For([]string{WORKER, SERVER}, func(key string) { push(kit.Format(":%s=", key)) })
				}
			}},
			ctx.CMDS: {Name: "cmds space index* args", Help: "命令", Hand: func(m *ice.Message, arg ...string) {
				_route_toast(m, m.Option(SPACE), append([]string{m.Option(ctx.INDEX)}, kit.Split(m.Option(ctx.ARGS))...)...)
			}},
			cli.BUILD: {Name: "build space", Help: "构建", Hand: func(m *ice.Message, arg ...string) {
				_route_toast(m, m.Option(SPACE), cli.RUNTIME, UPGRADE)
				func() { defer ToastProcess(m)(); m.Sleep3s() }()
				m.SetAppend().Cmdy("", "travel")
			}},
			"travel": {Help: "遍历", Icon: "bi bi-card-list", Hand: func(m *ice.Message, arg ...string) {
				kit.For(kit.Split(m.OptionDefault(ice.MSG_FIELDS, mdb.Config(m, mdb.FIELD))), func(key string) {
					switch key {
					case mdb.TIME:
						m.Push(key, ice.Info.Make.Time)
					case mdb.TYPE:
						m.Push(key, ice.Info.NodeType)
					case nfs.MODULE:
						m.Push(key, ice.Info.Make.Module)
					case nfs.VERSION:
						m.Push(key, ice.Info.Make.Versions())
					case "commitTime":
						m.Push(key, ice.Info.Make.When)
					case "compileTime":
						m.Push(key, ice.Info.Make.Time)
					case "bootTime":
						m.Push(key, m.Cmdx(cli.RUNTIME, "boot.time"))
					case "md5":
						m.Push(key, ice.Info.Hash)
					case nfs.SIZE:
						if m.Option(ice.MSG_USERPOD) == "" {
							defer ToastProcess(m)()
						}
						var stats runtime.MemStats
						runtime.ReadMemStats(&stats)
						m.Push(key, kit.Format("%s/%s/%s", kit.FmtSize(int64(stats.Sys)), ice.Info.Size, m.Cmdx(nfs.DIR, nfs.SIZE)))
					case nfs.PATH:
						m.Push(key, kit.Path(""))
					case tcp.HOSTNAME:
						m.Push(key, ice.Info.Hostname)
					default:
						m.Push(key, "")
					}
				})
				defer m.ProcessRefresh()
				PushPodCmd(m, "", m.ActionKey())
				m.Table(func(value ice.Maps) { kit.If(value[SPACE], func() { mdb.HashCreate(m.Spawn(), kit.Simple(value)) }) })
			}},
			"diagram": {Help: "导图", Icon: "bi bi-diagram-3", Hand: func(m *ice.Message, arg ...string) {
				ctx.DisplayStorySpide(m.Cmdy(""), nfs.DIR_ROOT, ice.Info.NodeName, mdb.FIELD, SPACE, lex.SPLIT, nfs.PT, ctx.ACTION, ice.MAIN)
			}},
			mdb.PRUNES: &ice.Action{Name: "prunes status=offline", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd("", func(value ice.Maps) {
					kit.If(value[mdb.STATUS] == OFFLINE, func() { mdb.HashRemove(m, SPACE, value[SPACE]) })
				})
				m.ProcessRefresh()
			}},
		}, mdb.HashAction(mdb.SHORT, SPACE, mdb.FIELD, "time,space,type,module,version,commitTime,compileTime,bootTime,md5,size,path,hostname", mdb.SORT, "type,space", html.CHECKBOX, ice.TRUE, mdb.ACTION, ice.MAIN)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) > 1 {
				_route_match(m, arg[0], func(value ice.Maps, i int, list []ice.Maps) {
					_route_push(m, value[SPACE], m.Cmd(SPACE, value[SPACE], arg[1:]))
				})
			} else if mdb.HashSelect(m, arg...); len(arg) > 0 {
				m.EchoIFrame(m.MergePod(arg[0]))
			} else {
				m.OptionFields("")
				list := m.CmdMap(SPACE, mdb.NAME)
				mem, disk, stat := 0, 0, map[string]int{}
				m.Table(func(value ice.Maps) {
					disk += kit.Int(kit.Select("", kit.Split(value[nfs.SIZE], nfs.PS), 2))
					mem += kit.Int(kit.Select("", kit.Split(value[nfs.SIZE], nfs.PS), 0))
					if _, ok := list[value[SPACE]]; ok {
						m.Push(mdb.STATUS, ONLINE)
						stat[ONLINE]++
					} else {
						m.Push(mdb.STATUS, OFFLINE)
						stat[OFFLINE]++
					}
				}).Sort("status,space", ice.STR_R, ice.STR).StatusTimeCount(stat, nfs.SIZE, kit.Format("%s/%s", kit.FmtSize(mem), kit.FmtSize(disk))).Options(ice.MSG_ACTION, "")
			}
		}},
	})
}
