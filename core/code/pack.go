package code

import (
	ice "github.com/shylinux/icebergs"
	"github.com/shylinux/icebergs/base/nfs"
	kit "github.com/shylinux/toolkits"

	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func _pack_volcanos(m *ice.Message, pack *os.File) {
	m.Option(nfs.DIR_ROOT, "usr/volcanos")
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.FILE)

	for _, k := range []string{"favicon.ico", "index.html", "index.css", "index.js", "proto.js", "frame.js", "cache.js", "cache.css"} {
		what := ""
		if f, e := os.Open("usr/volcanos/" + k); e == nil {
			defer f.Close()
			if b, e := ioutil.ReadAll(f); e == nil {
				what = fmt.Sprintf("%v", b)
			}
		}
		if k == "index.html" {
			k = ""
		}
		what = strings.ReplaceAll(what, " ", ",")
		pack.WriteString(fmt.Sprintf(`        "%s": []byte{%v},`+"\n", "/"+k, what[1:len(what)-1]))
	}
	for _, k := range []string{"lib", "pane", "plugin"} {
		m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
			what := ""
			if f, e := os.Open("usr/volcanos/" + value["path"]); e == nil {
				defer f.Close()
				if b, e := ioutil.ReadAll(f); e == nil {
					what = fmt.Sprintf("%v", b)
				}
			}
			what = strings.ReplaceAll(what, " ", ",")
			pack.WriteString(fmt.Sprintf(`        "%s": []byte{%v},`+"\n", "/"+value["path"], what[1:len(what)-1]))
		})
	}
	pack.WriteString("\n")
}
func _pack_learning(m *ice.Message, pack *os.File) {
	m.Option(nfs.DIR_ROOT, "usr/learning")
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.FILE)

	m.Cmd(nfs.DIR, "./").Table(func(index int, value map[string]string, head []string) {
		what := ""
		if f, e := os.Open("usr/learning/" + value["path"]); e == nil {
			defer f.Close()
			if b, e := ioutil.ReadAll(f); e == nil {
				what = fmt.Sprintf("%v", b)
			}
		}
		what = strings.ReplaceAll(what, " ", ",")
		pack.WriteString(fmt.Sprintf(`        "%s": []byte{%v},`+"\n", "usr/learning/"+value["path"], what[1:len(what)-1]))
	})
	pack.WriteString("\n")
}
func _pack_icebergs(m *ice.Message, pack *os.File) {
	m.Option(nfs.DIR_ROOT, "usr/icebergs")
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.FILE)

	m.Cmd(nfs.DIR, "./").Table(func(index int, value map[string]string, head []string) {
		what := ""
		if strings.HasPrefix(value["path"], "pack") {
			return
		}
		if f, e := os.Open("usr/icebergs/" + value["path"]); e == nil {
			defer f.Close()
			if b, e := ioutil.ReadAll(f); e == nil {
				what = fmt.Sprintf("%v", b)
			}
		}
		if len(what) > 0 {
			what = strings.ReplaceAll(what, " ", ",")
			pack.WriteString(fmt.Sprintf(`        "%s": []byte{%v},`+"\n", "usr/icebergs/"+value["path"], what[1:len(what)-1]))
		}
	})
	pack.WriteString("\n")
}
func _pack_intshell(m *ice.Message, pack *os.File) {
	m.Option(nfs.DIR_ROOT, "usr/intshell")
	m.Option(nfs.DIR_DEEP, "true")
	m.Option(nfs.DIR_TYPE, nfs.FILE)

	m.Cmd(nfs.DIR, "./").Table(func(index int, value map[string]string, head []string) {
		if strings.HasPrefix(value["path"], "pluged") {
			return
		}
		what := ""
		if f, e := os.Open("usr/intshell/" + value["path"]); e != nil {
			return
		} else {
			defer f.Close()
			if b, e := ioutil.ReadAll(f); e != nil {
				return
			} else {
				what = fmt.Sprintf("%v", b)
			}
		}
		what = strings.ReplaceAll(what, " ", ",")
		pack.WriteString(fmt.Sprintf(`        "%s": []byte{%v},`+"\n", "usr/intshell/"+value["path"], what[1:len(what)-1]))
	})
}

const (
	WEBPACK = "webpack"
	BINPACK = "binpack"
)

func init() {
	Index.Merge(&ice.Context{
		Commands: map[string]*ice.Command{
			WEBPACK: {Name: "webpack", Help: "打包", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				m.Option(nfs.DIR_ROOT, "usr/volcanos")
				m.Option(nfs.DIR_DEEP, "true")
				m.Option(nfs.DIR_TYPE, nfs.FILE)

				js, p, e := kit.Create("usr/volcanos/cache.js")
				m.Assert(e)
				defer js.Close()
				m.Echo(p)

				css, _, e := kit.Create("usr/volcanos/cache.css")
				m.Assert(e)
				defer css.Close()

				for _, k := range []string{"lib", "pane", "plugin"} {
					m.Cmd(nfs.DIR, k).Table(func(index int, value map[string]string, head []string) {
						if strings.HasSuffix(value["path"], ".css") {
							js.WriteString(`Volcanos.meta.cache["` + path.Join("/", value["path"]) + "\"] = []\n")
							css.WriteString(m.Cmdx(nfs.CAT, "usr/volcanos/"+value["path"]))
						}
						if strings.HasSuffix(value["path"], ".js") {
							js.WriteString(`_can_name = "` + path.Join("/", value["path"]) + "\"\n")
							js.WriteString(m.Cmdx(nfs.CAT, "usr/volcanos/"+value["path"]))
						}
					})
				}
				for _, k := range []string{"frame.js"} {
					js.WriteString(`_can_name = "` + path.Join("/", k) + "\"\n")
					js.WriteString(m.Cmdx(nfs.CAT, "usr/volcanos/"+k))
				}
				js.WriteString(`_can_name = ""` + "\n")

				if f, _, e := kit.Create("usr/volcanos/cache.html"); m.Assert(e) {
					f.WriteString(fmt.Sprintf(`
<!DOCTYPE html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width,initial-scale=0.7,user-scalable=no">
    <title>volcanos</title>
    <link rel="shortcut icon" type="image/ico" href="favicon.ico">
    <style type="text/css">%s</style>
    <style type="text/css">%s</style>
</head>
<body>
<script>%s</script>
<script>%s</script>
<script>%s</script>
<script>%s</script>
<script>Volcanos.meta.webpack = true</script>
</body>
`,
						m.Cmdx(nfs.CAT, "usr/volcanos/cache.css"),
						m.Cmdx(nfs.CAT, "usr/volcanos/index.css"),

						m.Cmdx(nfs.CAT, "usr/volcanos/proto.js"),
						m.Cmdx(nfs.CAT, "usr/volcanos/cache.js"),
						m.Cmdx(nfs.CAT, "usr/volcanos/cache_data.js"),
						m.Cmdx(nfs.CAT, "usr/volcanos/index.js"),
					))
				}
			}},
			BINPACK: {Name: "binpack", Help: "打包", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
				pack, p, e := kit.Create("usr/icebergs/pack/binpack.go")
				m.Assert(e)
				defer pack.Close()

				pack.WriteString(`package pack` + "\n\n")
				pack.WriteString(`import "github.com/shylinux/icebergs"` + "\n\n")
				pack.WriteString(`func init() {` + "\n")
				pack.WriteString(`    ice.BinPack = map[string][]byte{` + "\n")

				_pack_volcanos(m, pack)
				_pack_learning(m, pack)
				_pack_icebergs(m, pack)
				_pack_intshell(m, pack)

				pack.WriteString(`    }` + "\n")
				pack.WriteString(`}` + "\n")
				m.Echo(p)
			}},
		},
	}, nil)

}
