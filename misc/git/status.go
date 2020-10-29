package git

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"path"
	"strings"
)

const STATUS = "status"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			REPOS: {Name: REPOS, Help: "仓库", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_NAME, kit.MDB_FIELD, "time,name,branch,last",
				"owner", "https://github.com/shylinux",
			)},
		},
		Commands: map[string]*ice.Command{
			STATUS: {Name: "status name auto submit compile pull", Help: "代码状态", Action: map[string]*ice.Action{
				"pull": {Name: "pull", Help: "下载", Hand: func(m *ice.Message, arg ...string) {
					m.Option(cli.PROGRESS_CB, func(cb func(name string, count, total int)) {
						count, total := 0, len(m.Confm(REPOS, kit.MDB_HASH))
						m.Richs(REPOS, nil, kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
							value = kit.GetMeta(value)

							cb(kit.Format(value[kit.MDB_NAME]), count, total)
							m.Option(cli.CMD_DIR, value[kit.MDB_PATH])
							m.Echo(m.Cmdx(cli.SYSTEM, GIT, "pull"))
							count++
						})
						cb("", total, total)
					})
					m.Cmdy(cli.PROGRESS, mdb.CREATE)
				}},
				"compile": {Name: "compile", Help: "编译", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(cli.SYSTEM, "make")
				}},

				"add": {Name: "add", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					if strings.Contains(m.Option(kit.MDB_NAME), ":\\") {
						m.Option(cli.CMD_DIR, m.Option(kit.MDB_NAME))
					} else {
						m.Option(cli.CMD_DIR, path.Join("usr", m.Option(kit.MDB_NAME)))
					}
					m.Cmdy(cli.SYSTEM, GIT, "add", m.Option(kit.MDB_FILE))
				}},
				"submit": {Name: "submit action=opt,add comment=some", Help: "提交", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_NAME) == "" {
						return
					}
					if strings.Contains(m.Option(kit.MDB_NAME), ":\\") {
						m.Option(cli.CMD_DIR, m.Option(kit.MDB_NAME))
					} else {
						m.Option(cli.CMD_DIR, path.Join("usr", m.Option(kit.MDB_NAME)))
					}

					if arg[0] == "action" {
						m.Cmdy(cli.SYSTEM, GIT, "commit", "-am", kit.Select("opt some", arg[1]+" "+arg[3]))
					} else {
						m.Cmdy(cli.SYSTEM, GIT, "commit", "-am", kit.Select("opt some", strings.Join(arg, " ")))
					}
				}},
				"push": {Name: "push", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
					if m.Option(kit.MDB_NAME) == "" {
						return
					}
					if strings.Contains(m.Option(kit.MDB_NAME), ":\\") {
						m.Option(cli.CMD_DIR, m.Option(kit.MDB_NAME))
					} else {
						m.Option(cli.CMD_DIR, path.Join("usr", m.Option(kit.MDB_NAME)))
					}
					m.Cmdy(cli.SYSTEM, GIT, "push")
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Richs(REPOS, nil, kit.Select(kit.MDB_FOREACH, arg, 0), func(key string, value map[string]interface{}) {
					value = kit.GetMeta(value)

					if m.Option(cli.CMD_DIR, value[kit.MDB_PATH]); len(arg) > 0 {
						m.Echo(m.Cmdx(cli.SYSTEM, GIT, "diff"))
						return // 更改详情
					}

					// 更改列表
					for _, v := range strings.Split(strings.TrimSpace(m.Cmdx(cli.SYSTEM, GIT, "status", "-sb")), "\n") {
						vs := strings.SplitN(strings.TrimSpace(v), " ", 2)
						m.Push("name", value[kit.MDB_NAME])
						m.Push("tags", vs[0])
						m.Push("file", vs[1])

						list := []string{}
						switch vs[0] {
						case "##":
							if strings.Contains(vs[1], "ahead") {
								list = append(list, "push")
							}
						default:
							if strings.Contains(vs[0], "??") {
								list = append(list, "add")
							} else {
								list = append(list, "submit")
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
