package code

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/util/bench"
)

func _bench_http(m *ice.Message, target string, arg ...string) {
	nconn := kit.Int64(kit.Select("10", m.Option(NCONN)))
	nreqs := kit.Int64(kit.Select("100", m.Option(NREQS)))
	m.Echo("nconn: %d total: %d\n", nconn, nreqs*nconn)

	list := []*http.Request{}
	for _, v := range strings.Split(target, ice.NL) {
		switch ls := kit.Split(v); ls[0] {
		case http.MethodPost: // POST,url,file
			if f, e := os.Open(ls[2]); m.Assert(e) {
				defer f.Close()

				if req, err := http.NewRequest(http.MethodPost, ls[1], f); m.Assert(err) {
					list = append(list, req)
				}
			}
		default:
			if req, err := http.NewRequest(http.MethodGet, v, nil); m.Assert(err) {
				list = append(list, req)
			}
		}
	}

	var ndata int64
	if s, e := bench.HTTP(nconn, nreqs, list, func(req *http.Request, res *http.Response) {
		n, _ := io.Copy(ioutil.Discard, res.Body)
		atomic.AddInt64(&ndata, n)
	}); m.Assert(e) {
		m.Echo("ndata: %s\n", kit.FmtSize(ndata))
		m.Echo(s.Show())
		m.ProcessInner()
	}
}
func _bench_redis(m *ice.Message, target string, arg ...string) {
}

const (
	HTTP  = "http"
	REDIS = "redis"
)
const (
	NCONN = "nconn"
	NREQS = "nreqs"
)
const BENCH = "bench"

func init() {
	Index.MergeCommands(ice.Commands{
		BENCH: {Name: "bench zone id auto insert", Help: "性能压测", Actions: ice.MergeAction(ice.Actions{
			mdb.INSERT: {Name: "insert zone=some type=http,redis name=demo text='http://localhost:9020' nconn=3 nreqs=10", Help: "添加"},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(mdb.TYPE) {
				case HTTP:
					_bench_http(m, m.Option(mdb.TEXT))
				case REDIS:
					_bench_redis(m, m.Option(mdb.TEXT))
				}
			}},
		}, mdb.ZoneAction(mdb.SHORT, mdb.ZONE, mdb.FIELD, "time,id,type,name,text,nconn,nreqs")), Hand: func(m *ice.Message, arg ...string) {
			mdb.ZoneSelect(m, arg...).PushAction(kit.Select(mdb.REMOVE, ice.RUN, len(arg) > 0))
		}},
	})
}
