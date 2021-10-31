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

func _bench_http(m *ice.Message, name, target string, arg ...string) {
	nconn := kit.Int64(kit.Select("10", m.Option(NCONN)))
	nreqs := kit.Int64(kit.Select("1000", m.Option(NREQS)))
	m.Echo("nconn: %d nreqs: %d\n", nconn, nreqs*nconn)

	list := []*http.Request{}
	for _, v := range strings.Split(target, ice.SP) {
		switch ls := kit.Split(v); ls[0] {
		case http.MethodPost:
			if f, e := os.Open(ls[2]); m.Assert(e) {
				defer f.Close()

				req, err := http.NewRequest(http.MethodPost, ls[1], f)
				m.Assert(err)
				list = append(list, req)
			}
		default:
			req, err := http.NewRequest(http.MethodGet, v, nil)
			m.Assert(err)
			list = append(list, req)
		}
	}

	var body int64
	s, e := bench.HTTP(nconn, nreqs, list, func(req *http.Request, res *http.Response) {
		n, _ := io.Copy(ioutil.Discard, res.Body)
		atomic.AddInt64(&body, n)
	})
	m.Assert(e)

	m.Echo(s.Show())
	m.Echo("body: %d\n", body)
	m.ProcessInner()
}
func _bench_redis(m *ice.Message, name, target string, arg ...string) {
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
	Index.Merge(&ice.Context{Configs: map[string]*ice.Config{
		BENCH: {Name: BENCH, Help: "性能压测", Value: kit.Data(
			kit.MDB_SHORT, kit.MDB_ZONE, kit.MDB_FIELD, "time,id,type,name,text,nconn,nreqs",
		)},
	}, Commands: map[string]*ice.Command{
		"/bench": {Name: "/bench cmd...", Help: "性能压测", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			m.Cmdy(m.Optionv("cmd"))
		}},
		BENCH: {Name: "bench zone id auto insert", Help: "性能压测", Action: ice.MergeAction(map[string]*ice.Action{
			mdb.INSERT: {Name: "insert zone=some type=http,redis name=demo text nconn=3 nreqs=10", Help: "添加"},
			ice.RUN: {Name: "run", Help: "执行", Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(kit.MDB_TYPE) {
				case HTTP:
					_bench_http(m, m.Option(kit.MDB_NAME), m.Option(kit.MDB_TEXT))
				case REDIS:
					_bench_redis(m, m.Option(kit.MDB_NAME), m.Option(kit.MDB_TEXT))
				}
			}},
		}, mdb.ZoneAction()), Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			mdb.ZoneSelect(m, arg...)
			m.PushAction(kit.Select(mdb.REMOVE, ice.RUN, len(arg) > 0))
		}},
	}})
}
