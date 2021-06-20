package gdb

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/ctx"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

const (
	TASK_HASH = "task.hash"
)
const (
	INNER = "inner"
)
const ROUTINE = "routine"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ROUTINE: {Name: ROUTINE, Help: "协程池", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			ROUTINE: {Name: "routine hash auto prunes", Help: "协程池", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create fileline status", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, ROUTINE, "", mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, ROUTINE, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, ROUTINE, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Option(mdb.FIELDS, "time,hash,status,fileline")
					m.Cmdy(mdb.PRUNES, ROUTINE, "", mdb.HASH, kit.MDB_STATUS, cli.ERROR)
					m.Cmdy(mdb.PRUNES, ROUTINE, "", mdb.HASH, kit.MDB_STATUS, cli.STOP)
				}},

				INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
					switch kit.Select("", arg, 0) {
					case kit.SSH_RUN:
						m.Cmdy(INNER, arg[1:])
					default:
						ls := kit.Split(m.Option("fileline"), ":")
						m.ProcessField(INNER, kit.SSH_RUN)
						m.Option(kit.SSH_ARG, kit.Format([]string{path.Dir(ls[0]), path.Base(ls[0]), ls[1]}))
						m.Cmdy(ctx.COMMAND, INNER)
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg) == 0, "time,hash,status,fileline")
				m.Cmdy(mdb.SELECT, ROUTINE, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(INNER, mdb.REMOVE)
			}},
		},
	})
}
