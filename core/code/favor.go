package code

import (
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

const FAVOR = "favor"

func init() {
	Index.MergeCommands(ice.Commands{
		FAVOR: {Name: "favor zone id auto insert test page", Help: "收藏夹", Actions: ice.MergeAction(ice.Actions{
			mdb.INSERT: {Name: "insert zone=数据结构 type=go name=hi text=hello path file line", Help: "添加"},
			INNER: {Name: "inner", Help: "源码", Hand: func(m *ice.Message, arg ...string) {
				ctx.ProcessCommand(m, INNER, m.OptionSplit("path,file,line"), arg...)
			}},
			"test": {Name: "test zone=hi count=10", Help: "测试", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, m.OptionSimple(mdb.ZONE))
				for i := 0; i < kit.Int(m.Option(mdb.COUNT)); i++ {
					m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.ZONE, m.Option(mdb.ZONE), mdb.NAME, i)
				}
				m.Conf(m.PrefixKey(), kit.Keys(mdb.HASH, kit.Hashs(m.Option(mdb.ZONE)), kit.Keym(mdb.LIMIT)), 20)
				m.Conf(m.PrefixKey(), kit.Keys(mdb.HASH, kit.Hashs(m.Option(mdb.ZONE)), kit.Keym(mdb.LEAST)), 10)
				m.Conf(m.PrefixKey(), kit.Keys(mdb.HASH, kit.Hashs(m.Option(mdb.ZONE)), kit.Keym(mdb.FSIZE)), 500)
				m.Cmdy(nfs.DIR, "var/data/", kit.Dict(nfs.DIR_DEEP, true))
				m.Echo(kit.Formats(m.Confv(m.PrefixKey())))
			}},
		}, mdb.ZoneAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,type,name,text,path,file,line")), Hand: func(m *ice.Message, arg ...string) {
			mdb.ZoneSelectPage(m, arg...).PushAction(kit.Select(mdb.REMOVE, INNER, len(arg) > 0))
		}},
	})
}
