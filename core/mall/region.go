package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const (
	AREA       = "area"
	POPULATION = "population"
	GDP        = "gdp"
)
const REGION = "region"

func init() {
	Index.MergeCommands(ice.Commands{
		REGION: {Help: "地区", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(AREA, "面积(平方公里)", POPULATION, "人口(万人)", GDP, "产值(亿元)")),
		), Actions: ice.MergeActions(mdb.ExportHashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,gdp,population,area,text")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).SortIntR(GDP).StatusTimeCount(m.Stats(GDP, POPULATION, AREA))
		}},
	})
}
