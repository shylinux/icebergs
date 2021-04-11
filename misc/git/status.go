package git

import (
	"strings"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"
)

const (
	COMPILE = "compile"
	COMMIT  = "commit"
	DIFF    = "diff"
	MAKE    = "make"
)
const STATUS = "status"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		STATUS: {Name: "status name auto pull compile create commit", Help: "代码状态", Action: map[string]*ice.Action{
			PULL: {Name: "pull", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.PROGRESS, mdb.CREATE, func(update func(name string, count, total int)) {
					count, total := 0, len(m.Confm(REPOS, kit.MDB_HASH))
					m.Richs(REPOS, nil, kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
						value = kit.GetMeta(value)
						update(kit.Format(value[kit.MDB_NAME]), count, total)

						m.Option(cli.CMD_DIR, value[kit.MDB_PATH])
						m.Cmd(cli.SYSTEM, GIT, PULL)
						count++
					})
					update("", total, total)
				})
			}},
			COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(cli.SYSTEM, MAKE)
			}},
			mdb.CREATE: {Name: "create main=src/main.go@key name=hi@key from=usr/icebergs/misc/bash/bash.go@key", Help: "模块", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.AUTOGEN, mdb.CREATE, arg)
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(code.AUTOGEN, mdb.INPUTS, arg)
			}},

			ADD: {Name: "add", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Option(cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME)))
				m.Cmdy(cli.SYSTEM, GIT, ADD, m.Option(kit.MDB_FILE))
			}},
			COMMIT: {Name: "commit action=opt,add comment=some", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(kit.MDB_NAME) == "" {
					return
				}

				if m.Option(cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME))); arg[0] == kit.MDB_ACTION {
					m.Cmdy(cli.SYSTEM, GIT, COMMIT, "-am", kit.Select("opt some", arg[1]+" "+arg[3]))
				} else {
					m.Cmdy(cli.SYSTEM, GIT, COMMIT, "-am", kit.Select("opt some", strings.Join(arg, " ")))
				}
				m.Option(ice.MSG_PROCESS, ice.PROCESS_REFRESH)
			}},
			PUSH: {Name: "push", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(kit.MDB_NAME) == "" {
					return
				}

				if strings.Contains(m.Option(kit.MDB_NAME), ":\\") {
					m.Option(cli.CMD_DIR, m.Option(kit.MDB_NAME))
				} else {
					m.Option(cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME)))
				}
				m.Cmdy(cli.SYSTEM, GIT, PUSH)
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Richs(REPOS, nil, kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
				value = kit.GetMeta(value)
				if m.Option(cli.CMD_DIR, value[kit.MDB_PATH]); len(arg) > 0 {
					m.Echo(m.Cmdx(cli.SYSTEM, GIT, DIFF))
					return // 更改详情
				}

				// 更改列表
				for _, v := range strings.Split(strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, STATUS, "-sb")), "\n") {
					vs := strings.SplitN(strings.TrimSpace(v), " ", 2)
					m.Push(kit.MDB_NAME, value[kit.MDB_NAME])
					m.Push(kit.MDB_TYPE, vs[0])
					m.Push(kit.MDB_FILE, vs[1])

					list := []string{}
					switch vs[0] {
					case "##":
						if strings.Contains(vs[1], "ahead") {
							list = append(list, PUSH)
						}
					default:
						if strings.Contains(vs[0], "??") {
							list = append(list, ADD)
						} else {
							list = append(list, COMMIT)
						}
					}
					m.PushButton(strings.Join(list, ","))
				}
			})
			m.Sort(kit.MDB_NAME)
		}},
	}})
}
