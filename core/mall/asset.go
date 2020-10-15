package mall

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
)

func _asset_create(m *ice.Message, account string) {
	if msg := m.Cmd(mdb.SELECT, ASSET, "", mdb.HASH, ACCOUNT, account); len(msg.Appendv(kit.MDB_HASH)) == 0 {
		m.Conf(ASSET, kit.Keys(m.Option(ice.MSG_DOMAIN), kit.MDB_META, kit.MDB_SHORT), ACCOUNT)
		m.Cmdy(mdb.INSERT, ASSET, "", mdb.HASH, ACCOUNT, account)
	}
}
func _asset_insert(m *ice.Message, account string, arg ...string) {
	m.Richs(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN)), account, func(key string, value map[string]interface{}) {
		for i := 0; i < len(arg)-1; i += 2 {
			if arg[i] == "amount" {
				kit.Value(value, "meta.amount", kit.Int(kit.Value(value, "meta.amount"))+kit.Int(arg[i+1]))
			}
		}
		id := m.Grow(ASSET, kit.Keys(kit.MDB_HASH, m.Optionv(ice.MSG_DOMAIN), kit.MDB_HASH, key), kit.Dict(
			kit.MDB_EXTRA, kit.Dict(),
			arg,
		))
		m.Log_INSERT(ACCOUNT, account, kit.MDB_ID, id, arg[0], arg[1])
		m.Echo("%d", id)
	})
}

const ASSET = "asset"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			ASSET: {Name: ASSET, Help: "资产", Value: kit.Data(kit.MDB_SHORT, ACCOUNT)},
		},
		Commands: map[string]*ice.Command{
			ASSET: {Name: "asset account id auto insert export import", Help: "资产", Action: map[string]*ice.Action{
				mdb.INSERT: {Name: "insert account type=spend,trans,bonus amount name text", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_asset_create(m, arg[1])
					_asset_insert(m, arg[1], arg[2:]...)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_asset_modify(m, m.Option(ACCOUNT), m.Option(kit.MDB_ID), arg[0], arg[1])
				}},
				mdb.EXPORT: {Name: "export file", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					_asset_export(m, m.Option(kit.MDB_FILE))
				}},
				mdb.IMPORT: {Name: "import file", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					_asset_import(m, m.Option(kit.MDB_FILE))
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					_asset_inputs(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_asset_list(m, kit.Select("", arg, 0), kit.Select("", arg, 1))
			}},
		},
	}, nil)
}
