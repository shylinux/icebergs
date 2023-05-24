package cli

import (
	"os"
	"runtime"
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
	utime       int64
	stime       int64
	vmsize      int64
	vmrss       int64
	user        int64
	sys         int64
	idle        int64
	total       int64
	free        int64
	rx          int64
	tx          int64
	established int64
	time_wait   int64
}

func newprocstat(m *ice.Message) (stat procstat) {
	if runtime.GOOS != LINUX {
		return
	}
	m.Option(ice.MSG_USERROLE, aaa.ROOT)
	if ls := kit.Split(m.Cmdx(nfs.CAT, kit.Format("/proc/%d/stat", os.Getpid())), " ()"); len(ls) > 0 {
		stat = procstat{utime: kit.Int64(ls[13]), stime: kit.Int64(ls[14]), vmsize: kit.Int64(ls[22]), vmrss: kit.Int64(ls[23]) * 4096}
	}
	if ls := kit.Split(kit.Select("", strings.Split(m.Cmdx(nfs.CAT, "/proc/stat"), lex.NL), 1)); len(ls) > 0 {
		stat.user = kit.Int64(ls[1])
		stat.sys = kit.Int64(ls[3])
		stat.idle = kit.Int64(ls[4])
	}
	for _, line := range strings.Split(strings.TrimSpace(m.Cmdx(nfs.CAT, "/proc/meminfo")), lex.NL) {
		switch ls := kit.Split(line, ": "); ls[0] {
		case "MemTotal":
			stat.total = kit.Int64(ls[1]) * 1024
		case "MemFree":
			stat.free = kit.Int64(ls[1]) * 1024
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(m.Cmdx(nfs.CAT, "/proc/net/dev")), lex.NL)[2:] {
		if ls := kit.Split(line, ": "); ls[0] != "lo" {
			stat.rx += kit.Int64(ls[1])
			stat.tx += kit.Int64(ls[9])
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(m.Cmdx(nfs.CAT, "/proc/net/tcp")), lex.NL)[1:] {
		switch ls := kit.Split(line, ": "); ls[5] {
		case "01":
			stat.established++
		case "06":
			stat.time_wait++
		}
	}
	return
}

func init() {
	var last procstat
	Index.MergeCommands(ice.Commands{
		"procstat": {Name: "procstat id list page", Actions: ice.MergeActions(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) { last = newprocstat(m) }},
			mdb.INSERT: {Name: "insert", Hand: func(m *ice.Message, arg ...string) {
				stat := newprocstat(m)
				total := stat.user - last.user + stat.sys - last.sys + stat.idle - last.idle
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.LIST,
					"utime", (stat.utime-last.utime+stat.stime-last.stime)*1000/total, "vmrss", stat.vmrss*1000/stat.total,
					"user", (stat.user-last.user+stat.sys-last.sys)*1000/total, "idle", (stat.idle-last.idle)*1000/total, "free", stat.free*1000/stat.total,
					"rx", (stat.rx-last.rx)*1000/20000000, "tx", (stat.tx-last.tx)*1000/20000000, "established", stat.established, "time_wait", stat.time_wait,
				)
				last = stat
			}},
		}, mdb.PageListAction(mdb.LIMIT, "720", mdb.LEAST, "360", mdb.FIELD, "time,id,utime,vmrss,user,idle,free,rx,tx,established,time_wait")), Hand: func(m *ice.Message, arg ...string) {
			m.OptionDefault(mdb.CACHE_LIMIT, "360")
			if mdb.PageListSelect(m, arg...); (len(arg) == 0 || arg[0] == "") && m.Length() > 0 {
				m.SortInt(mdb.ID).Display("/plugin/story/trend.js", ice.VIEW, "折线图", "min", "0", "max", "1000", COLOR, "yellow,cyan,red,green,blue,purple,purple")
				m.Status("from", m.Append(mdb.TIME), "span", kit.FmtDuration(time.Duration(kit.Time(m.Time())-kit.Time(m.Append(mdb.TIME)))), m.AppendSimple(mdb.Config(m, mdb.FIELD)), "cursor", "0")
			}
		}},
	})
}
