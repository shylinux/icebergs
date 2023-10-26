package chat

import (
	"net/http"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	kit "shylinux.com/x/toolkits"
)

const (
	LATITUDE  = "latitude"
	LONGITUDE = "longitude"
)

const LOCATION = "location"

func init() {
	get := func(m *ice.Message, api string, arg ...ice.Any) string {
		return kit.Format(mdb.Cache(m, kit.Join(kit.Simple(api, arg)), func() ice.Any {
			res := kit.UnMarshal(m.Cmdx(web.SPIDE, LOCATION, web.SPIDE_RAW, http.MethodGet, api, mdb.KEY, mdb.Config(m, web.TOKEN), arg))
			m.Warn(kit.Format(kit.Value(res, mdb.STATUS)) != "0", kit.Format(res))
			m.Debug("what %v %v", api, kit.Formats(res))
			return res
		}))
	}
	Index.MergeCommands(ice.Commands{
		LOCATION: {Help: "地图", Icon: "Maps.png", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, LOCATION, "https://apis.map.qq.com/ws/")
			}},
			"explore": {Help: "周边", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(get(m, "place/v1/explore", m.OptionSimple("keyword,boundary,page_index")))
			}},
			"search": {Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(get(m, "place/v1/search", m.OptionSimple("keyword,boundary,page_index")))
			}},
			"direction": {Help: "导航", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(get(m, "direction/v1/"+m.Option(mdb.TYPE)+nfs.PS, m.OptionSimple("from,to")))
			}},
			"district": {Help: "地区", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(get(m, "district/v1/getchildren", m.OptionSimple(mdb.ID)))
			}},
			FAVOR_INPUTS: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(arg[0] == mdb.TYPE, func() { m.Push(arg[0], LOCATION) })
			}},
			FAVOR_TABLES: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == LOCATION, func() { m.PushButton(kit.Dict(LOCATION, "地图")) })
			}},
			FAVOR_ACTION: {Hand: func(m *ice.Message, arg ...string) {
				kit.If(m.Option(mdb.TYPE) == LOCATION, func() { ctx.ProcessField(m, m.PrefixKey(), []string{m.Option(mdb.TEXT)}, arg...) })
			}},
		}, FavorAction(), mdb.ExportHashAction(mdb.FIELD, "time,hash,type,name,text,latitude,longitude,extra", nfs.SCRIPT, "https://map.qq.com/api/gljs?v=1.exp")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, kit.Slice(arg, 0, 1)...)
			// m.Option(LOCATION, m.Cmdx(web.SERVE, tcp.HOST))
			// m.Option(LOCATION, get(m, "location/v1/ip", aaa.IP, m.Option(ice.MSG_USERIP)))
			web.PushPodCmd(m, "", arg...)
			ctx.DisplayLocal(m.Options(nfs.SCRIPT, kit.MergeURL(mdb.Config(m, nfs.SCRIPT), mdb.KEY, mdb.Config(m, web.TOKEN))), "")
			ctx.Toolkit(m, "")
		}},
	})
}
