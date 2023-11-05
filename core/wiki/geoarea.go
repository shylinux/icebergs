package wiki

import (
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	"shylinux.com/x/icebergs/base/web"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const GEOAREA = "geoarea"

func init() {
	Index.MergeCommands(ice.Commands{
		GEOAREA: {Name: "geoarea path auto", Help: "地区", Actions: ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(web.SPIDE, mdb.CREATE, GEOAREA, "https://geo.datav.aliyun.com/areas_v3/bound/")
			}},
			nfs.PS: {Hand: func(m *ice.Message, arg ...string) {
				p := path.Join(ice.USR_GEOAREA, path.Join(arg...))
				kit.If(!nfs.Exists(m, p), func() { m.Cmd(web.SPIDE, GEOAREA, web.SPIDE_SAVE, p, arg) })
				m.RenderDownload(p)
			}},
			"china": {Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, m.PrefixKey(), func() {
					m.Push(ctx.STYLE, html.FLOAT)
				}, arg...)
			}},
			"city": {Hand: func(m *ice.Message, arg ...string) {
				stat := map[string]int{}
				lead := map[string]string{}
				list := map[string][]string{}
				m.Cmdy(nfs.CAT, "usr/geoarea/city.txt", func(ls []string, text string) {
					if len(ls) < 2 {
						return
					}
					for _, k := range []string{
						"自治区", "自治州", "自治县", "自治旗", "盟", "旗",
						"特别行政区", "特别行政区",
						"地区",
						"林区",
						"特区",
						"省", "市", "县", "区",
					} {
						if strings.HasSuffix(ls[1], k) {
							stat[k]++
							break
						}
					}
					if strings.HasSuffix(ls[0], "0000") {
						lead[ls[0][0:2]] = ls[1]
						kit.If(strings.HasSuffix(ls[1], "市"), func() { stat["直辖市"]++ })
						stat["省级单位"]++ // 34 = 4 直辖市 23 省 5 自治区 2 特别行政区
					} else if strings.HasSuffix(ls[0], "00") {
						list[lead[ls[0][0:2]]] = append(list[lead[ls[0][0:2]]], ls[1])
						kit.If(strings.HasSuffix(ls[1], "市"), func() { stat["地级市"]++ })
						stat["地级单位"]++ // 333 = 293 地级市 30 自治州 3 盟 7 地区
					} else {
						// list[lead[ls[0][0:2]]] = append(list[lead[ls[0][0:2]]], ls[1])
						kit.If(strings.HasSuffix(ls[1], "市"), func() { stat["县级市"]++ })
						stat["县级单位"]++ // 2842 = 388 县级市 1312 县 117 自治县 3 自治旗 49 旗 5 林区 1 特区 967 市辖区
					}
				})
				for k, v := range list {
					m.Push(mdb.NAME, k)
					m.Push(mdb.VALUE, len(v))
					m.Push(mdb.LIST, strings.Join(v, ","))
				}
				m.SortIntR(mdb.VALUE).StatusTimeCount(stat)
				m.Action("china")
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			ctx.DisplayStoryChina(m.Options(mdb.NAME, "中国", nfs.PATH, kit.Select("", arg, 0)))
			m.Action("city")
		}},
	})

}
