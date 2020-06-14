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
)

func _bench_list(m *ice.Message, zone string, id string, field ...interface{}) {
	m.RichList(BENCH, zone, id, field)
}

func _bench_show(m *ice.Message, nconn, nreq int64, list []*http.Request) {
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
func _bench_create(m *ice.Message) {
}
func _bench_insert(m *ice.Message) {
}

func init() {
	Index.Merge(&ice.Context{
		Configs: map[string]*ice.Config{
			BENCH: {Name: "bench", Help: "性能压测", Value: kit.Data(
				kit.MDB_SHORT, kit.MDB_ZONE,
			)},
		},
		Commands: map[string]*ice.Command{
			BENCH: {Name: "bench list nconn nreq", Help: "性能压测", Action: map[string]*ice.Action{
				"show": {Name: "show", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					list := []*http.Request{}
					target := kit.Select("http://localhost:9020/", arg, 0)
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
					m.Echo("%s\n", target)
					_bench_show(m, kit.Int64(kit.Select("3", arg, 1)), kit.Int64(kit.Select("100", arg, 2)), list)
				}},
				"create": {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
					_bench_create(m)

				}},
				"insert": {Name: "insert", Help: "插入", Hand: func(m *ice.Message, arg ...string) {
					_bench_insert(m)
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				_bench_list(m, BENCH, kit.Select("", arg, 0), kit.Select("", arg, 1))
			}},
			"/bench": {Name: "/bench cmd...", Help: "性能压测", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Cmdy(m.Optionv("cmd"))
			}},
		},
	}, nil)
}
