package chat

import (
	"sync"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	LATITUDE  = "latitude"
	LONGITUDE = "longitude"
)

const LOCATION = "location"

func init() {
	location := sync.Map{}
	cache := func(m *ice.Message, key string, load func() string) ice.Any {
		if current, ok := location.Load(key); ok {
			m.Debug("read cache %v", key)
			return current
		}
		current := load()
		location.Store(key, current)
		m.Debug("load cache %v %v", key, current)
		return current
	}
	get := func(m *ice.Message, api string, arg ...ice.Any) string {
		return kit.Format(cache(m, kit.Join(kit.Simple(api, arg)), func() string {
			return m.Cmdx(web.SPIDE_GET, "https://apis.map.qq.com/ws/"+api, mdb.KEY, m.Config(aaa.TOKEN), arg)
		}))
	}

	Index.MergeCommands(ice.Commands{
		LOCATION: {Name: "location hash auto", Help: "地理位置", Actions: ice.MergeActions(ice.Actions{
			"explore": {Name: "explore", Help: "周边", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(get(m, "place/v1/explore", m.OptionSimple("boundary,page_index")))
			}},
			"search": {Name: "search", Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if arg[0] == "*" {
					return
				}
				m.Echo(get(m, "place/v1/search", m.OptionSimple("keyword,boundary,page_index")))
			}},
			"direction": {Name: "direction", Help: "导航", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(get(m, "direction/v1/"+m.Option(mdb.TYPE)+ice.PS, m.OptionSimple("from,to")))
			}},
			"district": {Name: "district", Help: "地区", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(get(m, "district/v1/getchildren", m.OptionSimple(mdb.ID)))
			}},
		}, mdb.HashAction(mdb.FIELD, "time,hash,type,name,text,latitude,longitude,extra"), ctx.CmdAction()), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, kit.Slice(arg, 0, 1)...)
			ctx.DisplayLocal(m, "", m.ConfigSimple(aaa.TOKEN))
			m.Option(LOCATION, get(m, "location/v1/ip", aaa.IP, m.Option(ice.MSG_USERIP)))
		}},
	})
}
