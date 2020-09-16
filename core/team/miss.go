package team

const MISS = "miss"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			MISS: {Name: "miss", Help: "miss", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE)},
		},
		Commands: map[string]*ice.Command{},
	}, nil)
}
