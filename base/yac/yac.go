package yac

import (
	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _yac_load(m *ice.Message) {
	m.Richs(m.Prefix(MATRIX), "", kit.MDB_FOREACH, func(key string, value map[string]interface{}) {
		value = kit.GetMeta(value)

		mat := NewMatrix(m, kit.Int(kit.Select("32", value[NLANG])), kit.Int(kit.Select("32", value[NCELL])))
		m.Grows(m.Prefix(MATRIX), kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
			page := mat.index(m, NPAGE, kit.Format(value[NPAGE]))
			hash := mat.index(m, NHASH, kit.Format(value[NHASH]))
			if mat.mat[page] == nil {
				mat.mat[page] = make([]*State, mat.ncell)
			}

			mat.train(m, page, hash, kit.Simple(value[kit.MDB_TEXT]), 1)
		})
		value[MATRIX] = mat
	})
}

const YAC = "yac"

var Index = &ice.Context{Name: YAC, Help: "语法模块",
	Commands: map[string]*ice.Command{
		ice.CTX_INIT: {Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			_yac_load(m.Load())
		}},
		ice.CTX_EXIT: {Hand: func(m *ice.Message, c *ice.Context, key string, arg ...string) {
			m.Save()
		}},
	},
}

func init() { ice.Index.Register(Index, nil) }
