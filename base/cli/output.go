package cli

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"io"
)

func _cli_progress(m *ice.Message, cb interface{}) {
	m.Option(mdb.FIELDS, "time,name,step,count,total")
	if m.Option(ice.MSG_PROCESS, "_progress"); m.Option("_progress") != "" {
		m.Cmdy(mdb.SELECT, PROGRESS, "", mdb.HASH, kit.MDB_HASH, m.Option("_progress"))
		if len(m.Appendv(kit.MDB_TIME)) > 0 {
			return
		}
	}

	h := m.Cmdx(mdb.INSERT, PROGRESS, "", mdb.HASH, "step", 0, "count", 0, "total", 1)
	m.Cmdy(mdb.SELECT, PROGRESS, "", mdb.HASH, kit.MDB_HASH, m.Option("_progress", h))

	m.Go(func() {
		switch cb := cb.(type) {
		case func(cb func(name string, count, total int)):
			cb(func(name string, count, total int) {
				m.Cmd(mdb.MODIFY, PROGRESS, "", mdb.HASH, kit.MDB_HASH, h, "name", name,
					"step", count*100/total, "count", count, "total", total)
			})
		}
	})
}
func Follow(m *ice.Message, action string, cb func()) bool {
	m.Option(ice.MSG_PROCESS, "_follow")
	if m.Option(mdb.CACHE_ACTION, action); m.Option(mdb.CACHE_HASH) != "" {
		m.Cmdy(OUTPUT, m.Option(mdb.CACHE_HASH))
		m.Sort(kit.MDB_ID).Table(func(index int, value map[string]string, head []string) {
			m.Option(mdb.CACHE_BEGIN, value[kit.MDB_ID])
			m.Echo(value[kit.SSH_RES])
		})

		if len(m.Resultv()) == 0 && m.Conf(OUTPUT, kit.Keys(kit.MDB_HASH, m.Option(mdb.CACHE_HASH), kit.MDB_META, kit.MDB_STATUS)) == STOP {
			m.Option(mdb.CACHE_STATUS, STOP)
		}
		return true
	}
	m.Cmdy(OUTPUT, mdb.CREATE, kit.MDB_NAME, m.Option(kit.MDB_LINK))
	m.Option(mdb.CACHE_HASH, m.Result())
	m.Option(mdb.CACHE_BEGIN, 1)
	m.Set(ice.MSG_RESULT)
	m.Go(cb)
	return false
}

const (
	PROGRESS_CB = "progress.cb"
)

const PROGRESS = "progress"

const OUTPUT = "output"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			OUTPUT:   {Name: OUTPUT, Help: "输出", Value: kit.Data()},
			PROGRESS: {Name: PROGRESS, Help: "进度", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			PROGRESS: {Name: "progress hash auto", Help: "进度", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_cli_progress(m, m.Optionv(PROGRESS_CB))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,name,step,count,total", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, PROGRESS, "", mdb.HASH, kit.MDB_HASH, arg)
				m.SortInt("step")
			}},

			OUTPUT: {Name: "output hash id auto", Help: "输出", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create cmd", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					h := m.Cmdx(mdb.INSERT, OUTPUT, "", mdb.HASH, kit.MDB_STATUS, Status.Start, arg)
					r, w := io.Pipe()
					m.Go(func() {
						buf := make([]byte, 1024)
						for {
							if n, e := r.Read(buf); e != nil || n == 0 {
								break
							} else {
								m.Grow(OUTPUT, kit.Keys(kit.MDB_HASH, h), kit.Dict(
									kit.SSH_RES, string(buf[:n]),
								))
							}
						}
						m.Cmd(mdb.MODIFY, OUTPUT, "", mdb.HASH, kit.MDB_HASH, h, kit.MDB_STATUS, Status.Stop)
					})
					m.Option(OUTPUT, w)
					m.Echo(h)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, OUTPUT, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, OUTPUT, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH))
				}},
				mdb.PRUNES: {Name: "prunes", Help: "清理", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.PRUNES, OUTPUT, "", mdb.HASH, kit.MDB_STATUS, Status.Stop)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Option(mdb.FIELDS, "time,hash,status,name")
					m.Cmdy(mdb.SELECT, OUTPUT, "", mdb.HASH)
					m.PushAction(mdb.REMOVE)
					return
				}

				m.Option(ice.MSG_CONTROL, "_page")
				m.Option(mdb.FIELDS, kit.Select("time,id,res", mdb.DETAIL, len(arg) > 1))
				m.Cmdy(mdb.SELECT, OUTPUT, kit.Keys(kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, arg[1:])
				m.Sort(kit.MDB_ID)
			}},
		},
	})
}
