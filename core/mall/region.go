package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const (
	AREA       = "area"
	POPULATION = "population"
	GDP        = "gdp"
	CITY       = "city"
)
const REGION = "region"

func init() {
	Index.MergeCommands(ice.Commands{
		REGION: {Help: "地区", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(AREA, "面积(平方公里)", POPULATION, "人口(万人)", GDP, "产值(亿元)")),
		), Actions: ice.MergeActions(ice.Actions{
			GDP: {Help: "产值", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, m.PrefixKey(), func() {
					m.Push(ctx.DISPLAY, "/plugin/story/china.js?title=全国产值分布(亿元)&field=gdp&style=float")
				}, arg...)
			}},
			POPULATION: {Help: "人口", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, m.PrefixKey(), func() {
					m.Push(ctx.DISPLAY, "/plugin/story/china.js?title=全国人口分布(万人)&field=population&style=float")
				}, arg...)
			}},
			AREA: {Help: "土地", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, m.PrefixKey(), func() {
					m.Push(ctx.DISPLAY, "/plugin/story/china.js?title=全国土地分布(平方公里)&field=area&style=float")
				}, arg...)
			}},
			CITY: {Help: "本地", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessField(m, m.PrefixKey(), func() {
					m.Push(ctx.DISPLAY, "/plugin/story/china.js?title=深圳资源分布&field=area&style=float&path=440300")
				}, arg...)
			}},
		}, mdb.ExportHashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,gdp,population,area,text")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).SortIntR(GDP).Action(mdb.CREATE, GDP, POPULATION, AREA, CITY).StatusTimeCount(m.Stats(GDP, POPULATION, AREA))
		}},
	})
}
