package web

import (
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/web/html"
	kit "shylinux.com/x/toolkits"
)

func _count_stat(m *ice.Message, arg ...string) map[string]int {
	stat := map[string]int{}
	m.Table(func(value ice.Maps) {
		count := kit.Int(value[mdb.COUNT])
		stat[mdb.TOTAL] += count
		for _, agent := range []string{"美国", "电信", "联通", "移动", "阿里云", "腾讯云"} {
			if strings.Contains(value[aaa.LOCATION], agent) {
				stat[agent] += count
				break
			}
		}
		for _, agent := range []string{"GoModuleMirror", "Go-http-client", "git", "compatible"} {
			if strings.Contains(value[mdb.TEXT], agent) {
				stat[agent] += count
				return
			}
		}
		for _, agent := range html.AgentList {
			if strings.Contains(value[mdb.TEXT], agent) {
				stat[agent] += count
				break
			}
		}
		for _, agent := range html.SystemList {
			if strings.Contains(value[mdb.TEXT], agent) {
				stat[agent] += count
				break
			}
		}
	})
	return stat
}

const COUNT = "count"

func init() {
	Index.MergeCommands(ice.Commands{
		COUNT: &ice.Command{Name: "count hash auto group valid location filter", Help: "计数器", Meta: kit.Dict(
			ice.CTX_TRANS, kit.Dict(html.INPUT, kit.Dict(aaa.LOCATION, "地理位置")),
		), Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create type name text", Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelectUpdate(m, mdb.HashCreate(m), func(value ice.Map) { value[mdb.COUNT] = kit.Int(value[mdb.COUNT]) + 1 })
			}},
			mdb.VALID: {Hand: func(m *ice.Message, arg ...string) {
				mdb.HashSelect(m.Spawn(), arg...).Table(func(value ice.Maps) {
					if !strings.HasPrefix(value[mdb.TEXT], html.Mozilla) {
						return
					} else if count := kit.Int(value[mdb.COUNT]); count < 1 {
						return
					} else {
						m.Push("", value, kit.Split(mdb.Config(m, mdb.FIELD)))
					}
				})
				m.StatusTimeCount(_count_stat(m))
			}},
			mdb.GROUP: {Hand: func(m *ice.Message, arg ...string) {
				count := map[string]int{}
				list := map[string]map[string]string{}
				m.Cmd("", mdb.VALID).Table(func(value ice.Maps) {
					count[value[aaa.LOCATION]] += kit.Int(value[mdb.COUNT])
					list[value[aaa.LOCATION]] = value
				})
				stat := map[string]int{}
				for _, v := range list {
					func() {
						for _, agent := range []string{"美国", "电信", "联通", "移动", "阿里云", "腾讯云", "北京市", "香港"} {
							if strings.Contains(v[aaa.LOCATION], agent) {
								stat[agent] += kit.Int(v[mdb.COUNT])
								return
							}
						}
						m.Push("", v, kit.Split(mdb.Config(m, mdb.FIELD)))
					}()
				}
				m.StatusTimeCount(stat)
			}},
			aaa.LOCATION: {Hand: func(m *ice.Message, arg ...string) {
				GoToast(mdb.HashSelects(m).Sort(mdb.COUNT, ice.INT_R), func(toast func(string, int, int)) []string {
					m.Table(func(value ice.Maps, index, total int) {
						if value[aaa.LOCATION] == "" {
							location := kit.Format(kit.Value(SpideGet(m, "http://opendata.baidu.com/api.php?co=&resource_id=6006&oe=utf8", "query", value[mdb.NAME]), "data.0.location"))
							mdb.HashModify(m, mdb.HASH, value[mdb.HASH], aaa.LOCATION, location)
							toast(kit.Select(value[mdb.NAME], location), index, total)
							m.Sleep300ms()
						}
					})
					return nil
				})
			}},
		}, mdb.HashAction(mdb.LIMIT, 1000, mdb.LEAST, 500, mdb.SHORT, "type,name", mdb.FIELD, "time,hash,count,location,type,name,text")), Hand: func(m *ice.Message, arg ...string) {
			mdb.HashSelect(m, arg...).Sort(mdb.TIME, ice.STR_R).StatusTimeCount(_count_stat(m))
		}},
	})
}

func Count(m *ice.Message, arg ...string) *ice.Message {
	kit.If(len(arg) > 0 && arg[0] == "", func() { arg[0] = ctx.ShortCmd(m.PrefixKey()) })
	kit.If(len(arg) > 1 && arg[1] == "", func() { arg[1] = m.ActionKey() })
	m.Cmd(COUNT, mdb.CREATE, arg, kit.Dict(ice.LOG_DISABLE, ice.TRUE))
	return m
}
