package team

import (
	ice "github.com/shylinux/icebergs"
	kit "github.com/shylinux/toolkits"
)

const MISS = "miss"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			MISS: {Name: "miss", Help: "miss", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE)},
		},
		Commands: map[string]*ice.Command{},
	}, nil)
}
