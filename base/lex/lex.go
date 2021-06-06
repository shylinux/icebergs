package lex

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const LEX = "lex"

var Index = &ice.Context{Name: LEX, Help: "词法模块",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			return
			m.Load()
			m.Richs(m.Prefix(MATRIX), "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
				value = kit.GetMeta(value)

				mat := NewMatrix(m, kit.Int(kit.Select("32", value[NLANG])), kit.Int(kit.Select("256", value[NCELL])))
				m.Grows(m.Prefix(MATRIX), kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
					mat.Train(m, kit.Format(value[NPAGE]), kit.Format(value[NHASH]), kit.Format(value[kit.MDB_TEXT]))
				})
				value[MATRIX] = mat
			})
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			return
			m.Save()
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
