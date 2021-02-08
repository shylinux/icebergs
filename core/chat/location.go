package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"

	"math"
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
	ADDRESS   = "address"
	LATITUDE  = "latitude"
	LONGITUDE = "longitude"
)
const (
	GETLOCATION  = "getLocation"
	OPENLOCATION = "openLocation"
)

const LOCATION = "location"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			LOCATION: {Name: LOCATION, Help: "地理位置", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			LOCATION: {Name: "location hash auto getLocation", Help: "地理位置", Action: map[string]*ice.Action{
				OPENLOCATION: {Name: "openLocation", Help: "地图", Hand: func(m *ice.Message, arg ...string) {}},
				GETLOCATION: {Name: "getLocation", Help: "打卡", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, LOCATION, "", mdb.HASH, arg)
				}},
				mdb.CREATE: {Name: "create type=text name text latitude longitude", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, LOCATION, "", mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, LOCATION, "", mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, LOCATION, "", mdb.HASH, kit.MDB_TEXT, m.Option(kit.MDB_TEXT))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, LOCATION, "", mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, LOCATION, "", mdb.HASH)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INPUTS, LOCATION, "", mdb.HASH, arg)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,hash,type,name,text,longitude,latitude", mdb.DETAIL, len(arg) > 0))
				m.Cmdy(mdb.SELECT, cmd, "", mdb.HASH, kit.MDB_HASH, arg)
				m.PushAction(OPENLOCATION, mdb.REMOVE)
			}},
		},
	})
}
