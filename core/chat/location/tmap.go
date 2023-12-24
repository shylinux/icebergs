package location

import (
	"net/http"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/core/chat"
	kit "shylinux.com/x/toolkits"
)

const (
	SEARCH    = "search"
	EXPLORE   = "explore"
	CURRENT   = "current"
	DIRECTION = "direction"
)
const TMAP = "tmap"

func init() {
	get := func(m *ice.Message, api string, arg ...ice.Any) string {
		return kit.Format(mdb.Cache(m, kit.Join(kit.Simple(api, arg)), func() ice.Any {
			res := kit.UnMarshal(m.Cmdx(web.SPIDE, TMAP, web.SPIDE_RAW, http.MethodGet, api, mdb.KEY, mdb.Config(m, aaa.SECRET), arg))
			m.Warn(kit.Format(kit.Value(res, mdb.STATUS)) != "0", kit.Format(res))
			m.Debug("what %v %v", api, kit.Formats(res))
			return res
		}))
	}
	chat.Index.MergeCommands(ice.Commands{
		TMAP: {Help: "腾讯地图", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, TMAP, "https://apis.map.qq.com/ws/")
			}},
			DISTRICT: {Help: "地区", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(get(m, "district/v1/getchildren", m.OptionSimple(mdb.ID)))
			}},
			EXPLORE: {Help: "周边", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(get(m, "place/v1/explore", m.OptionSimple("keyword,boundary,page_index")))
			}},
			SEARCH: {Help: "搜索", Hand: func(m *ice.Message, arg ...string) {
				if m.Option("keyword") == "" {
					return
				}
				m.Echo(get(m, "place/v1/search", m.OptionSimple("keyword,boundary,page_index")))
			}},
			DIRECTION: {Help: "导航", Hand: func(m *ice.Message, arg ...string) {
				m.Echo(get(m, "direction/v1/"+m.Option(mdb.TYPE)+nfs.PS, m.OptionSimple("from,to")))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			// m.Option(LOCATION, m.Cmdx(web.SERVE, tcp.HOST))
			// m.Option(LOCATION, get(m, "location/v1/ip", aaa.IP, m.Option(ice.MSG_USERIP)))
			m.Display("", nfs.SCRIPT, mdb.Config(m, nfs.SCRIPT))
		}},
	})
}
