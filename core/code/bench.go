package code

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
	"shylinux.com/x/toolkits/util/bench"
)

func _bench_http(m *ice.Message, target string, arg ...string) {
	list := []*http.Request{}
	nconn := kit.Int64(kit.Select("10", m.Option(NCONN)))
	nreqs := kit.Int64(kit.Select("100", m.Option(NREQS)))
	m.Cmd(nfs.CAT, "", kit.Dict(nfs.CAT_CONTENT, target), func(ls []string, text string) {
		if len(ls) == 0 || strings.HasPrefix(text, "#") {
			return
		}
		switch ls[0] {
		case http.MethodPost:
			if f, e := nfs.OpenFile(m, ls[2]); m.Assert(e) {
				defer f.Close()
				if req, err := http.NewRequest(http.MethodPost, ls[1], f); m.Assert(err) {
					list = append(list, req)
				}
			}
		case http.MethodGet:
			ls = ls[1:]
			fallthrough
		default:
			if req, err := http.NewRequest(http.MethodGet, ls[0], nil); m.Assert(err) {
				list = append(list, req)
			}
		}
	})
	var ndata int64
	if s, e := bench.HTTP(m.FormatTaskMeta(), nconn, nreqs, list, func(req *http.Request, res *http.Response) {
		n, _ := io.Copy(ioutil.Discard, res.Body)
		atomic.AddInt64(&ndata, n)
	}); m.Assert(e) {
		m.Echo("nconn: %d total: %d ndata: %s\n", nconn, nreqs*nconn, kit.FmtSize(ndata)).Echo(s.Show()).ProcessInner()
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
		BENCH: {Help: "压测", Actions: ice.MergeActions(ice.Actions{
			mdb.INSERT: {Name: "insert zone*=demo type*=http,redis name=demo text*='http://localhost:9020/chat/cmd/web.chat.favor' nconn=10 nreqs=100"},
			cli.START: {Hand: func(m *ice.Message, arg ...string) {
				switch m.Option(mdb.TYPE) {
				case HTTP:
					_bench_http(m, m.Option(mdb.TEXT))
				case REDIS:
					_bench_redis(m, m.Option(mdb.TEXT))
				}
			}},
		}, mdb.ZoneAction(mdb.FIELDS, "time,id,type,name,text,nconn,nreqs")), Hand: func(m *ice.Message, arg ...string) {
			mdb.ZoneSelect(m, arg...).PushAction(kit.Select(cli.START, mdb.REMOVE, len(arg) == 0))
		}},
	})
}
