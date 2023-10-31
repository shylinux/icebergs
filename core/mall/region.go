package mall

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

const REGION = "region"

func init() {
	Index.MergeCommands(ice.Commands{
		REGION: {Help: "地区", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(
				"area", "面积", "population", "人口", "gdp", "产值",
			)),
		), Actions: ice.MergeActions(mdb.ExportHashAction(mdb.SHORT, mdb.NAME, mdb.FIELD, "time,name,text,area,population,gdp"))},
	})
}
