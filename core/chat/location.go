package chat

import (
	"math"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

func distance(lat1, long1, lat2, long2 float64) float64 {
	lat1 = lat1 * math.Pi / 180
	long1 = long1 * math.Pi / 180
	lat2 = lat2 * math.Pi / 180
	long2 = long2 * math.Pi / 180
	return 2 * 6371 * math.Asin(math.Sqrt(math.Pow(math.Sin(math.Abs(lat1-lat2)/2), 2)+math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(math.Abs(long1-long2)/2), 2)))
}
func _trans(arg []string, tr map[string]string) {
	for i := 0; i < len(arg)-1; i += 2 {
		arg[i] = kit.Select(arg[i], tr[arg[i]])
	}
}

const (
	LATITUDE  = "latitude"
	LONGITUDE = "longitude"
)
const (
	GETLOCATION  = "getLocation"
	OPENLOCATION = "openLocation"
)
const LOCATION = "location"

func init() {
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		LOCATION: {Name: LOCATION, Help: "地理位置", Value: kit.Data(
			mdb.SHORT, mdb.TEXT, mdb.FIELD, "time,hash,type,name,text,longitude,latitude",
		)},
	}, Commands: map[string]*ice.Command{
		LOCATION: {Name: "location hash auto getLocation", Help: "地理位置", Action: ice.MergeAction(map[string]*ice.Action{
			OPENLOCATION: {Name: "location", Help: "地图"},
			GETLOCATION:  {Name: "location create", Help: "打卡"},
			mdb.CREATE:   {Name: "create type=text name text latitude longitude", Help: "添加"},
		}, mdb.HashAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.HashSelect(m, arg...)
			m.PushAction(OPENLOCATION, mdb.REMOVE)
		}},
	}})
}
