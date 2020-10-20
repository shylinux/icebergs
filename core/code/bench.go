package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	kit "github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/util/bench"

	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

func _bench_http(m *ice.Message, kind, name, target string, arg ...string) {
	nconn := kit.Int64(kit.Select("10", m.Option(NCONN)))
	nreqs := kit.Int64(kit.Select("1000", m.Option(NREQS)))
	m.Echo("nconn: %d nreqs: %d\n", nconn, nreqs*nconn)

	list := []*http.Request{}
	for _, v := range strings.Split(target, ",") {
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
	m.Option(ice.MSG_PROCESS, "_inner")
}

const (
	NCONN = "nconn"
	NREQS = "nreqs"
)

const BENCH = "bench"

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			BENCH: {Name: BENCH, Help: "性能压测", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE)},
		},
		Commands: map[string]*ice.Command{
			BENCH: {Name: "bench zone=auto id=auto auto insert", Help: "性能压测", Action: map[string]*ice.Action{
				mdb.CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, BENCH, "", mdb.HASH, arg)
				}},
				mdb.INSERT: {Name: "insert zone type=http,redis name text nconn=3 nreqs=10", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.INSERT, BENCH, "", mdb.HASH, kit.MDB_ZONE, arg[1])
					m.Cmdy(mdb.INSERT, BENCH, _sub_key(m, m.Option(kit.MDB_ZONE)), mdb.LIST, arg[2:])
				}},
				mdb.MODIFY: {Name: "modify", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.MODIFY, BENCH, _sub_key(m, m.Option(kit.MDB_ZONE)), mdb.LIST, kit.MDB_ID, m.Option(kit.MDB_ID), arg)
				}},
				mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
					m.Cmdy(mdb.DELETE, BENCH, "", mdb.HASH, kit.MDB_ZONE, m.Option(kit.MDB_ZONE))
				}},

				cli.RUN: {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					switch m.Option(kit.MDB_TYPE) {
					case "http":
						_bench_http(m, m.Option(kit.MDB_TYPE), m.Option(kit.MDB_NAME), m.Option(kit.MDB_TEXT))
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(mdb.FIELDS, kit.Select("time,count,zone", kit.Select("time,id,type,name,text,nconn,nreqs", mdb.DETAIL, len(arg) > 1), len(arg) > 0))
				m.Cmdy(mdb.SELECT, BENCH, "", mdb.ZONE, arg)
				m.PushAction(kit.Select(mdb.REMOVE, cli.RUN, len(arg) > 0))
			}},

			"/bench": {Name: "/bench cmd...", Help: "性能压测", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(m.Optionv("cmd"))
			}},
		},
	}, nil)
}
