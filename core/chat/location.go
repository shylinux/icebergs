package chat

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/web"
	kit "github.com/shylinux/toolkits"

	"math"
	"net/url"
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

const LOCATION = "location"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			LOCATION: {Name: LOCATION, Help: "地理位置", Value: kit.Data(kit.MDB_SHORT, kit.MDB_TEXT)},
		},
		Commands: map[string]*ice.Command{
			LOCATION: {Name: "location text auto 添加@location", Help: "地理位置", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "insert type name address latitude longitude", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					_trans(arg, map[string]string{"address": "text"})
					m.Conf(LOCATION, kit.Keys(m.Option(ice.MSG_DOMAIN), kit.MDB_META, kit.MDB_SHORT), kit.MDB_TEXT)
					m.Cmdy(mdb.INSERT, LOCATION, m.Option(ice.MSG_DOMAIN), mdb.HASH, arg)
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, LOCATION, m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_HASH, m.Option(kit.MDB_HASH), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, LOCATION, m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_TEXT, m.Option(kit.MDB_TEXT))
				}},
				mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.EXPORT, m.Prefix(LOCATION), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				}},
				mdb.IMPORT: {Name: "import", Help: "导入", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.IMPORT, m.Prefix(LOCATION), m.Option(ice.MSG_DOMAIN), mdb.HASH)
				}},
				mdb.INPUTS: {Name: "inputs", Help: "补全", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INPUTS, LOCATION, m.Option(ice.MSG_DOMAIN), mdb.HASH, arg)
				}},

				mdb.RENDER: {Name: "render type name text", Help: "渲染", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.RENDER, web.RENDER.Frame, kit.Format(
						"https://map.baidu.com/search/%s/@12958750.085,4825785.55,16z?querytype=s&da_src=shareurl&wd=%s",
						arg[2], arg[2]))
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, "time,type,name,text,longitude,latitude")
				m.Cmdy(mdb.SELECT, LOCATION, m.Option(ice.MSG_DOMAIN), mdb.HASH, kit.MDB_HASH, arg)
				m.Table(func(index int, value map[string]string, head []string) {
					m.PushRender(kit.MDB_LINK, "a", "百度地图", kit.Format(
						"https://map.baidu.com/search/%s/@12958750.085,4825785.55,16z?querytype=s&da_src=shareurl&wd=%s",
						url.QueryEscape(kit.Format(value[kit.MDB_TEXT])),
						url.QueryEscape(kit.Format(value[kit.MDB_TEXT])),
					))
				})
				m.PushAction("删除")
			}},
		},
	}, nil)
}
