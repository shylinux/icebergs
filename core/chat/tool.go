package chat

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const STORM = "storm"
const TOOL = "tool"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		TOOL: {Name: "tool hash id auto insert create", Help: "工具", Action: map[string]*ice.Action{
			mdb.CREATE: {Name: "create type=public,protected,private name=hi text=hello", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m, TOOL), mdb.HASH, arg)
			}},
			mdb.INSERT: {Name: "insert hash pod ctx cmd help", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.INSERT, RIVER, _river_key(m, TOOL, m.OptionSimple(kit.MDB_HASH)), mdb.LIST, arg[2:])
			}},
			mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(kit.MDB_ID) == "" {
					m.Cmdy(mdb.MODIFY, RIVER, _river_key(m, TOOL), mdb.HASH, m.OptionSimple(kit.MDB_HASH), arg)
				} else {
					m.Cmdy(mdb.MODIFY, RIVER, _river_key(m, TOOL, m.OptionSimple(kit.MDB_HASH)), mdb.LIST, m.OptionSimple(kit.MDB_ID), arg)
				}
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				m.Cmdy(mdb.DELETE, RIVER, _river_key(m, TOOL), mdb.HASH, m.OptionSimple(kit.MDB_HASH))
			}},
			mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(kit.MDB_ID) != "" {
					msg := m.Cmd(TOOL, m.Option(kit.MDB_HASH), m.Option(kit.MDB_ID))
					cmd := kit.Keys(msg.Append(cli.CTX), msg.Append(cli.CMD))

					_action_domain(m, cmd, m.Option(kit.MDB_HASH))
					m.Cmdy(m.Space(msg.Append(cli.POD)), cmd, mdb.EXPORT)
				}
			}},
			mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
				if m.Option(kit.MDB_ID) != "" {
					msg := m.Cmd(TOOL, m.Option(kit.MDB_HASH), m.Option(kit.MDB_ID))
					cmd := kit.Keys(msg.Append(cli.CTX), msg.Append(cli.CMD))

					_action_domain(m, cmd, m.Option(kit.MDB_HASH))
					m.Cmdy(m.Space(msg.Append(cli.POD)), cmd, mdb.IMPORT)
				}
			}},
			mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
				if cli.Inputs(m, arg[0]) {
					return
				}

				switch arg[0] {
				case kit.MDB_HASH:
					m.Cmdy(TOOL, ice.OptionFields("hash,name"))
				}
			}},
		}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 0 {
				m.OptionFields("time,hash,type,name,count")
				m.Cmdy(mdb.SELECT, RIVER, _river_key(m, TOOL), mdb.HASH)
				m.PushAction(mdb.REMOVE)
				m.Sort(kit.MDB_NAME)
				return // 应用列表
			}

			m.OptionFields("time,id,pod,ctx,cmd,arg,display,style")
			msg := m.Cmd(mdb.SELECT, RIVER, _river_key(m, TOOL, kit.MDB_HASH, arg[0]), mdb.LIST, kit.MDB_ID, kit.Select("", arg, 1))
			if len(msg.Appendv(cli.CMD)) == 0 && len(arg) > 1 {
				msg.Push(cli.CMD, arg[1])
			}

			if len(arg) > 2 && arg[2] == cli.RUN {
				m.Cmdy(m.Space(msg.Append(cli.POD)), kit.Keys(msg.Append(cli.CTX), msg.Append(cli.CMD)), arg[3:])
				return // 执行命令
			}

			if m.Copy(msg); len(arg) < 2 {
				m.PushAction(mdb.EXPORT, mdb.IMPORT)
				return // 命令列表
			}

			// 命令插件
			m.ProcessField(arg[0], arg[1], cli.RUN)
			m.Table(func(index int, value map[string]string, head []string) {
				m.Cmdy(m.Space(value[cli.POD]), ctx.CONTEXT, value[cli.CTX], ctx.COMMAND, value[cli.CMD])
			})
		}},
	}})
}
