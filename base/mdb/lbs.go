package mdb

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"

	"math"
)

func distance(lat1, long1, lat2, long2 float64) float64 {
	lat1 = lat1 * math.Pi / 180
	long1 = long1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	long2 = long2 * math.Pi / 180
	return 2 * 6371 * math.Asin(math.Sqrt(math.Pow(math.Sin(math.Abs(lat1-lat2)/2), 2)+math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(math.Abs(long1-long2)/2), 2)))
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			"location": {Name: "location", Help: "定位", Value: kit.Data(kit.MDB_SHORT, "name")},
		},
		Commands: map[string]*ice.Command{
			"location": {Name: "location", Help: "location", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 {
					m.Grows("location", nil, "", "", func(index int, value map[string]interface{}) {
						m.Push("", value)
					})
					return
				}
				if len(arg) == 1 {
					m.Richs("location", nil, arg[0], func(key string, value map[string]interface{}) {
						m.Info("what %v", value)
						m.Push("detail", value)
					})
					return
				}
				if len(arg) == 2 {
					m.Richs("aaa.location", nil, "*", func(key string, value map[string]interface{}) {
						m.Push("name", value["name"])
						m.Push("distance", kit.Int(distance(
							float64(kit.Int(arg[0]))/100000,
							float64(kit.Int(arg[1]))/100000,
							float64(kit.Int(value["latitude"]))/100000,
							float64(kit.Int(value["longitude"]))/100000,
						)*1000))
					})
					m.Sort("distance", "int")
					return
				}

				data := m.Richs("location", nil, arg[0], nil)
				if data != nil {
					data["count"] = kit.Int(data["count"]) + 1
				} else {
					data = kit.Dict("name", arg[0], "address", arg[1], "latitude", arg[2], "longitude", arg[3], "count", 1)
					m.Rich("location", nil, data)
				}
				m.Grow("location", nil, data)
			}},
		},
	}, nil)
}
