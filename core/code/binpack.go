package code

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/cli"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _binpack_file(m *ice.Message, w io.Writer, arg ...string) {
	if strings.HasPrefix(arg[0], "usr/volcanos/publish/") && !strings.HasSuffix(arg[0], "/proto.js") {
		return
	}
	if strings.HasPrefix(arg[0], "usr/volcanos/page/") && !strings.Contains(arg[0], "/cache.") {
		fmt.Fprintf(w, "        \"%s\": \"%s\",\n", kit.Select(arg[0], arg, 1), "")
		return
	}
	switch path.Base(arg[0]) {
	case "go.mod", "go.sum":
		return
	}
	switch arg[0] {
	case ice.SRC_BINPACK_GO, ice.SRC_VERSION_GO, ice.ETC_LOCAL_SHY:
		return
	}
	if f, e := nfs.OpenFile(m, arg[0]); !m.Warn(e, ice.ErrNotFound, arg[0]) {
		defer f.Close()
		if b, e := ioutil.ReadAll(f); !m.Warn(e, ice.ErrNotValid, arg[0]) {
			fmt.Fprintf(w, "        \"%s\": \"%s\",\n", kit.Select(arg[0], arg, 1), base64.StdEncoding.EncodeToString(b))
		}
	}
}
func _binpack_dir(m *ice.Message, w io.Writer, dir string) {
	nfs.DirDeepAll(m, dir, nfs.PWD, func(value ice.Maps) { _binpack_file(m, w, path.Join(dir, value[nfs.PATH])) })
}
func _binpack_all(m *ice.Message) {
	nfs.OptionFiles(m, nfs.DiskFile)
	if w, p, e := nfs.CreateFile(m, ice.SRC_BINPACK_GO); m.Assert(e) {
		defer w.Close()
		defer m.Echo(p)
		fmt.Fprintln(w, nfs.Template(m, ice.SRC_BINPACK_GO))
		defer fmt.Fprintln(w, nfs.Template(m, "binpack_end.go"))
		for _, p := range []string{ice.USR_VOLCANOS, ice.USR_INTSHELL, ice.SRC} {
			_binpack_dir(m, w, p)
		}
		for _, p := range []string{ice.ETC_MISS_SH, ice.ETC_INIT_SHY, ice.ETC_EXIT_SHY, ice.README_MD, ice.MAKEFILE, ice.LICENSE} {
			_binpack_file(m, w, p)
		}
		list, cache := map[string]bool{}, kit.Select(ice.USR_REQUIRE, m.Cmdx(cli.SYSTEM, GO, "env", "GOMODCACHE"))
		for k := range ice.Info.File {
			switch ls := strings.Split(k, ice.PS); ls[2] {
			case ice.SRC:
			case ice.USR:
				list[path.Join(kit.Slice(ls, 2, -1)...)] = true
			default:
				list[path.Join(cache, path.Join(kit.Slice(ls, 2, -1)...))] = true
			}
		}
		for _, p := range kit.SortedKey(list) {
			m.Cmd(nfs.DIR, p, nfs.PATH, kit.Dict(nfs.DIR_ROOT, nfs.PWD, nfs.DIR_REG, kit.ExtReg("(sh|shy|py|js|css|html)"))).Tables(func(value ice.Maps) {
				if strings.Contains(value[nfs.PATH], "/go/pkg/mod/") {
					_binpack_file(m, w, value[nfs.PATH], ice.USR_REQUIRE+strings.Split(value[nfs.PATH], "/go/pkg/mod/")[1])
				} else {
					_binpack_file(m, w, value[nfs.PATH])
				}
			})
		}
		mdb.HashSelects(m).Sort(nfs.PATH).Tables(func(value ice.Maps) {
			if s, e := nfs.StatFile(m, value[nfs.PATH]); e == nil {
				if s.IsDir() {
					_binpack_dir(m, w, value[nfs.PATH])
				} else {
					_binpack_file(m, w, value[nfs.PATH])
				}
			}
		})
	}
}

const BINPACK = "binpack"

func init() {
	Index.MergeCommands(ice.Commands{
		BINPACK: {Name: "binpack path auto create insert", Help: "打包", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Hand: func(m *ice.Message, arg ...string) { _binpack_all(m) }},
			mdb.INSERT: {Name: "insert path*", Hand: func(m *ice.Message, arg ...string) { mdb.HashCreate(m) }},
		}, mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path"))},
	})
}
