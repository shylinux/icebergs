package cli

import (
	"strings"
	"time"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/aaa"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

type procstat struct {
	user int
	sys  int
	idle int
	io   int

	total     int
	free      int
	available int

	rx int
	tx int
}

func newprocstat(m *ice.Message) (stat procstat) {
	if ls := kit.Split(kit.Select("", strings.Split(m.Cmdx(nfs.CAT, "/proc/stat"), lex.NL), 1)); len(ls) > 0 {
		stat = procstat{
			user: kit.Int(ls[1]),
			sys:  kit.Int(ls[3]),
			idle: kit.Int(ls[4]),
			io:   kit.Int(ls[5]),
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(m.Cmdx(nfs.CAT, "/proc/meminfo")), lex.NL) {
		switch ls := kit.Split(line, ": "); ls[0] {
		case "MemTotal":
			stat.total = kit.Int(ls[1]) * 1024
		case "MemFree":
			stat.free = kit.Int(ls[1]) * 1024
		case "MemAvailable":
			stat.available = kit.Int(ls[1]) * 1024
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(m.Cmdx(nfs.CAT, "/proc/net/dev")), lex.NL)[2:] {
		ls := kit.Split(line, ": ")
		if ls[0] == "eth0" {
			stat.rx = kit.Int(ls[1])
			stat.tx = kit.Int(ls[9])
		}
	}
	return
}

func init() {
	var last procstat
	Index.MergeCommands(ice.Commands{
		"procstat": {Name: "procstat id auto page insert", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				aaa.White(m, "/proc/net/dev")
				aaa.White(m, "/proc/meminfo")
				aaa.White(m, "/proc/stat")
				last = newprocstat(m)
			}},
			mdb.INSERT: {Name: "insert", Hand: func(m *ice.Message, arg ...string) {
				stat := newprocstat(m)
				total := stat.user - last.user + stat.sys - last.sys + stat.idle - last.idle + stat.io - last.io
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.LIST,
					"user", (stat.user-last.user)*1000/total, "sys", (stat.sys-last.sys)*1000/total,
					"idle", (stat.idle-last.idle)*1000/total, "io", (stat.io-stat.io)*1000/total,
					"free", stat.free*1000/stat.total, "available", stat.available*1000/stat.total,
					"rx", (stat.rx-last.rx)*1000/10000000, "tx", (stat.tx-last.tx)*1000/10000000,
				)
				last = stat
			}},
		}, mdb.PageListAction(mdb.FIELD, "time,id,user,sys,idle,io,free,available,rx,tx")), Hand: func(m *ice.Message, arg ...string) {
			m.OptionDefault(mdb.CACHE_LIMIT, "1000")
			mdb.PageListSelect(m, arg...)
			m.SortInt(mdb.ID).Display("/plugin/story/trend.js", ice.VIEW, "折线图", "min", "0", "max", "1000", mdb.FIELD, "user,sys,idle,free,available,tx,rx", COLOR, "red,yellow,green,blue,cyan,purple")
			m.Status("from", m.Append(mdb.TIME), "span", kit.FmtDuration(time.Duration(kit.Time(m.Time())-kit.Time(m.Append(mdb.TIME)))), m.AppendSimple("time,user,sys,idle,free,available,tx,rx"), "cursor", "0")
		}},
	})
}
