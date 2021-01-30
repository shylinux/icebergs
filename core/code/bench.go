package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/cli"
	"github.com/shylinux/icebergs/base/mdb"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"github.com/shylinux/toolkits/util/bench"

	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync/atomic"
)

func _bench_http(m *ice.Message, name, target string, arg ...string) {
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
	m.Option(ice.MSG_PROCESS, ice.PROCESS_INNER)
}
func _bench_redis(m *ice.Message, name, target string, arg ...string) {
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
			BENCH: {Name: "bench zone id auto insert", Help: "性能压测", Action: map[string]*ice.Action{
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
						_bench_http(m, m.Option(kit.MDB_NAME), m.Option(kit.MDB_TEXT))
					case "redis":
						_bench_redis(m, m.Option(kit.MDB_NAME), m.Option(kit.MDB_TEXT))
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

			"test": {Name: "test path func auto run case", Help: "测试用例", Action: map[string]*ice.Action{
				"run": {Name: "run", Help: "运行", Hand: func(m *ice.Message, arg ...string) {
					cli.Follow(m, "run", func() {
						m.Option(cli.CMD_DIR, kit.Select(path.Dir(arg[0]), arg[0], strings.HasSuffix(arg[0], "/")))
						m.Cmdy(cli.SYSTEM, "go", "test", "./", "-v", "-run="+arg[1])
					})
				}},
				"case": {Name: "case", Help: "用例", Hand: func(m *ice.Message, arg ...string) {
					msg := m.Spawn()
					if strings.HasSuffix(arg[0], "/") {
						msg.Option(cli.CMD_DIR, arg[0])
						msg.Split(msg.Cmdx(cli.SYSTEM, "grep", "-r", "func Test.*(", "./"), "file:line", ":", "\n")
						msg.Table(func(index int, value map[string]string, head []string) {
							if strings.HasPrefix(strings.TrimSpace(value["line"]), "//") {
								return
							}
							ls := kit.Split(value["line"], " (", " (", " (")
							m.Push("file", value["file"])
							m.Push("func", strings.TrimPrefix(ls[1], "Test"))
						})
					} else {
						for _, line := range kit.Split(m.Cmdx(cli.SYSTEM, "grep", "^func Test.*(", arg[0]), "\n", "\n", "\n") {
							ls := kit.Split(line, " (", " (", " (")
							m.Push("func", strings.TrimPrefix(ls[1], "Test"))
						}
					}
				}},
			}, Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				if len(arg) == 0 || arg[0] == "" {
					m.Cmdy(nfs.DIR, "./")
					return
				}
				if len(arg) == 1 {
					if strings.HasSuffix(arg[0], "/") {
						m.Cmdy(nfs.DIR, arg[0])
					} else {
						for _, line := range kit.Split(m.Cmdx(cli.SYSTEM, "grep", "^func Test.*(", arg[0]), "\n", "\n", "\n") {
							ls := kit.Split(line, " (", " (", " (")
							m.Push("func", strings.TrimPrefix(ls[1], "Test"))
						}
					}
					return
				}

				m.Option(cli.CMD_DIR, kit.Select(path.Dir(arg[0]), arg[0], strings.HasSuffix(arg[0], "/")))
				m.Cmdy(cli.SYSTEM, "go", "test", "./", "-v", "-run="+arg[1])
			}},
		},
	})
}
