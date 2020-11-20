package git

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/core/code"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
)

const (
	PULL    = "pull"
	COMPILE = "compile"
	ADD     = "add"
	COMMIT  = "commit"
	PUSH    = "push"
)
const STATUS = "status"

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			STATUS: {Name: "status name auto commit compile pull", Help: "代码状态", Action: map[string]*ice.Action{
				PULL: {Name: "pull", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.PROGRESS_CB, func(cb func(name string, count, total int)) {
						count, total := 0, len(m.Confm(REPOS, kit.MDB_HASH))
						m.Richs(REPOS, nil, kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
							value = kit.GetMeta(value)

							cb(kit.Format(value[kit.MDB_NAME]), count, total)
							m.Option(cli.CMD_DIR, value[kit.MDB_PATH])
							m.Cmd(cli.SYSTEM, GIT, PULL)
							count++
						})
						cb("", total, total)
					})
					m.Cmdy(cli.PROGRESS, mdb.CREATE)
				}},
				COMPILE: {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "make")
				}},
				ADD: {Name: "add", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME)))
					m.Cmdy(cli.SYSTEM, GIT, ADD, m.Option(kit.MDB_FILE))
				}},
				COMMIT: {Name: "commit action=opt,add comment=some", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_NAME) == "" {
						return
					}
					m.Option(cli.CMD_DIR, _repos_path(m.Option(kit.MDB_NAME)))

					if arg[0] == "action" {
						m.Cmdy(cli.SYSTEM, GIT, COMMIT, "-am", kit.Select("opt some", arg[1]+" "+arg[3]))
					} else {
						m.Cmdy(cli.SYSTEM, GIT, COMMIT, "-am", kit.Select("opt some", strings.Join(arg, " ")))
					}
				}},
				PUSH: {Name: "push", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_NAME) == "" {
						return
					}
					if strings.Contains(m.Option(kit.MDB_NAME), ":\\") {
						m.Option(cli.CMD_DIR, m.Option(kit.MDB_NAME))
					} else {
						m.Option(cli.CMD_DIR, path.Join("usr", m.Option(kit.MDB_NAME)))
					}
					m.Cmdy(cli.SYSTEM, GIT, PUSH)
				}},

				mdb.CREATE: {Name: "create main=src/main.go@key name=hi@key from=usr/icebergs/misc/bash/bash.go@key", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(code.AUTOGEN, mdb.CREATE, arg)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(code.AUTOGEN, mdb.INPUTS, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(REPOS, nil, kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
					value = kit.GetMeta(value)

					if m.Option(cli.CMD_DIR, value[kit.MDB_PATH]); len(arg) > 0 {
						m.Echo(m.Cmdx(cli.SYSTEM, GIT, "diff"))
						return // 更改详情
					}

					// 更改列表
					for _, v := range strings.Split(strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, STATUS, "-sb")), "\n") {
						vs := strings.SplitN(strings.TrimSpace(v), " ", 2)
						m.Push("name", value[kit.MDB_NAME])
						m.Push("tags", vs[0])
						m.Push("file", vs[1])

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
		},
	})
}