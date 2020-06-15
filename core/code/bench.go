package code

import (
	"github.com/shylinux/icebergs"
	"github.com/shylinux/toolkits"
	"github.com/shylinux/toolkits/util/bench"

	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

const (
	BENCH = "bench"
	NCONN = "nconn"
	NREQS = "nreqs"
)

func _bench_list(m *ice.Message, zone string, id string, field ...interface{}) {
	m.Richs(BENCH, nil, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		if zone = kit.Format(kit.Value(val, "meta.zone")); id == "" {
			m.Grows(BENCH, kit.Keys(kit.MDB_HASH, key), "", "", func(index int, value map[string]interface{}) {
				m.Push(kit.MDB_ZONE, zone)
				m.Push("操作", `<input type="button" value="运行">`)
				m.Push(zone, value, field...)
			})
			return
		}
		m.Grows(BENCH, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			m.Push("detail", value)
		})
	})
}
func _bench_show(m *ice.Message, nconn, nreq int64, list []*http.Request) {
	m.Info(NCONN, nconn, NREQS, nreq)
	nout, e := os.OpenFile("/dev/null", os.O_WRONLY, 0777)
	m.Assert(e)

	var body int64
	s, e := bench.HTTP(nconn, nreq, list, func(req *http.Request, res *http.Response) {
		n, _ := io.Copy(nout, res.Body)
		atomic.AddInt64(&body, n)
	})
	m.Assert(e)

	m.Echo(s.Show())
	m.Echo("body: %d\n", body)
}
func _bench_create(m *ice.Message, zone string, arg ...string) {
	if m.Richs(BENCH, nil, zone, nil) == nil {
		m.Rich(BENCH, nil, kit.Data(kit.MDB_ZONE, zone, arg))
		m.Log_CREATE(kit.MDB_ZONE, zone)
	}
}
func _bench_insert(m *ice.Message, zone, kind, name, text string, nconn, nreqs string, arg ...string) {
	m.Richs(BENCH, nil, zone, func(key string, value map[string]interface{}) {
		id := m.Grow(BENCH, kit.Keys(kit.MDB_HASH, key), kit.Dict(
			kit.MDB_TYPE, kind, kit.MDB_NAME, name, kit.MDB_TEXT, text,
			NCONN, nconn, NREQS, nreqs,
			kit.MDB_EXTRA, kit.Dict(arg),
		))
		m.Log_INSERT(kit.MDB_ZONE, zone, kit.MDB_ID, id, kit.MDB_TYPE, kind, kit.MDB_NAME, name)
		m.Echo("%d", id)
	})
}
func _bench_modify(m *ice.Message, zone, id, pro, set, old string) {
	m.Richs(BENCH, nil, kit.Select(kit.MDB_FOREACH, zone), func(key string, val map[string]interface{}) {
		m.Grows(BENCH, kit.Keys(kit.MDB_HASH, key), kit.MDB_ID, id, func(index int, value map[string]interface{}) {
			switch pro {
			case kit.MDB_ZONE, kit.MDB_ID, kit.MDB_TIME:
				m.Info("not allow %v", key)
			default:
				m.Log_MODIFY(kit.MDB_ZONE, zone, kit.MDB_ID, id, kit.MDB_KEY, pro, kit.MDB_VALUE, set, "old", old)
				kit.Value(value, pro, set)
			}
		})
	})
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			BENCH: {Name: "bench", Help: "性能压测", Value: kit.Data(kit.MDB_SHORT, kit.MDB_ZONE)},
		},
		Commands: map[string]*ice.Command{
			BENCH: {Name: "bench zone=auto id=auto auto", Help: "性能压测", Action: map[string]*ice.Action{
				kit.MDB_SHOW: {Name: "show", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					list := []*http.Request{}
					target := kit.Select(m.Option(kit.MDB_TEXT))
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
					m.Echo("%s \n", target)
					_bench_show(m, kit.Int64(kit.Select(m.Option(NCONN))), kit.Int64(kit.Select(m.Option(NREQS))), list)
				}},
				kit.MDB_MODIFY: {Name: "modify key value old", Help: "编辑", Hand: func(m *ice.Message, arg ...string) {
					_bench_modify(m, m.Option(kit.MDB_ZONE), m.Option(kit.MDB_ID), arg[0], arg[1], "")
				}},
				kit.MDB_INSERT: {Name: "insert zone type name text nconn nreqs", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
					_bench_insert(m, arg[0], arg[1], arg[2],
						kit.Select("http://localhost:9020/code/bench?cmd="+arg[2], arg, 3),
						kit.Select("3", arg, 4), kit.Select("10", arg, 5))
				}},
				kit.MDB_CREATE: {Name: "create zone", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_bench_create(m, arg[0])
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_bench_list(m, kit.Select(kit.MDB_FOREACH, arg, 0), kit.Select("", arg, 1))
			}},
			"/bench": {Name: "/bench cmd...", Help: "性能压测", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(m.Optionv("cmd"))
			}},
		},
	}, nil)
}
