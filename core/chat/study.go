package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

var _target_insert = kit.List(
	"_input", "text", "name", "type", "value", "@key",
	"_input", "text", "name", "name", "value", "@key",
	"_input", "text", "name", "text", "value", "@key",
)

const (
	STUDY  = "study"
	TARGET = "target"
	// ACTION = "action"
	ASSESS = "assess"
)

func init() {
	Index.Register(&ice.Context{Name: STUDY, Help: "study",
		Configs: map[string]*ice.Config{
			TARGET: {Name: "target", Help: "大纲", Value: kit.Data()},
			ACTION: {Name: "action", Help: "互动", Value: kit.Data()},
			ASSESS: {Name: "assess", Help: "评测", Value: kit.Data()},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Load() }},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) { m.Save() }},

			TARGET: {Name: "target hash=auto auto 添加:button 导出:button 导入:button", Help: "大纲", Meta: kit.Dict(
				"添加", _target_insert,
			), Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert [key value]...", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(TARGET), "", mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify key value", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(TARGET), "", mdb.HASH, "", m.Option("hash"), arg[0], arg[1])
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(TARGET), "", mdb.HASH, "", m.Option("hash"))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(TARGET), "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(TARGET), "", mdb.HASH)
				}},

				mdb.INPUTS: {Name: "inputs key value", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INPUTS, m.Prefix(TARGET), "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Fields(len(arg), "time,hash,type,name,text")
				m.Cmdy(mdb.SELECT, m.Prefix(TARGET), "", mdb.HASH, "", kit.Select(kit.MDB_FOREACH, arg, 0))
				if len(arg) == 0 {
					m.PushAction("备课", "学习", "测试", "删除")
				}
			}},
			ACTION: {Name: "action", Help: "互动", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			}},
			ASSESS: {Name: "assess", Help: "评测", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			}},
		},
	}, nil)
}
