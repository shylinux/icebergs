package chat

import (
	"path"

	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"
)

const MEET = "meet"
const (
	MISS = "miss"
	DATE = "date"
)

var _miss_select = "time,name,性别,年龄,籍贯,学历,职业,照片"
var _miss_insert = kit.List(
	"_input", "text", "name", "name",
	"_input", "text", "name", "性别",
	"_input", "text", "name", "年龄",
	"_input", "text", "name", "生日",
	"_input", "text", "name", "照片",
	"_input", "text", "name", "身高",
	"_input", "text", "name", "体重",
	"_input", "text", "name", "性格",
	"_input", "text", "name", "爱好",
	"_input", "text", "name", "籍贯",
	"_input", "text", "name", "户口",
	"_input", "text", "name", "学历",
	"_input", "text", "name", "学校",
	"_input", "text", "name", "职业",
	"_input", "text", "name", "年薪",
	"_input", "text", "name", "资产",
)

var _date_select = "time,id,name,地点,主题"
var _date_insert = kit.List(
	"_input", "text", "name", "name",
	"_input", "text", "name", "地点",
	"_input", "text", "name", "主题",
)

func init() {
	Index.Register(&ice.Context{Name: MEET, Help: "遇见",
		Configs: map[string]*ice.Config{
			MISS: {Name: MISS, Help: "miss", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
			DATE: {Name: DATE, Help: "date", Value: kit.Data(kit.MDB_SHORT, kit.MDB_NAME)},
		},
		Commands: map[string]*ice.Command{
			ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Load(MISS, DATE)
			}},
			ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Save(MISS, DATE)
			}},

			MISS: {Name: "miss name=auto auto 添加:button 导出:button 导入:button", Help: "miss", Meta: kit.Dict(
				"display", "", "insert", _miss_insert,
			), Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert [key value]...", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(MISS), "", mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify key value", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(MISS), "", mdb.HASH, kit.MDB_NAME, m.Option("name"), arg[0], arg[1])
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, m.Prefix(MISS), "", mdb.HASH, kit.MDB_NAME, m.Option("name"))
				}},
				mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(MISS), "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(MISS), "", mdb.HASH)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option("fields", _miss_select)
				if msg := m.Cmd(mdb.SELECT, m.Prefix(MISS), "", mdb.HASH, kit.MDB_NAME, kit.Select(kit.MDB_FOREACH, arg, 0)); len(arg) == 0 {
					msg.Table(func(index int, value map[string]string, head []string) {
						for _, k := range head {
							if k == "照片" {
								m.Push("照片", m.Cmdx(mdb.RENDER, web.RENDER.IMG, path.Join("/share/local", value["照片"])))
							} else {
								m.Push(k, value[k])
							}
						}
					})
					m.PushAction("删除")
				}
			}},

			DATE: {Name: "date name=auto auto 添加:button 导出:button 导入:button", Help: "date", Meta: kit.Dict(
				"display", "", "insert", _date_insert,
			), Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert [key value]...", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, m.Prefix(DATE), "", mdb.LIST, arg)
				}},
				mdb.MODIFY: {Name: "modify key value", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, m.Prefix(DATE), "", mdb.LIST, kit.MDB_ID, m.Option("id"), arg[0], arg[1])
				}},
				mdb.DELETE: {Name: "delete", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					// m.Cmdy(mdb.DELETE, m.Prefix(DATE), "", mdb.LIST, kit.MDB_NAME, m.Option("name"))
				}},
				mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(DATE), "", mdb.LIST)
				}},
				mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(DATE), "", mdb.LIST)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option("fields", _date_select)
				m.Cmdy(mdb.SELECT, m.Prefix(DATE), "", mdb.LIST)
			}},
		},
	}, nil)
}
