package code

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/lex"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _binpack_file(m *ice.Message, w io.Writer, arg ...string) {
	if kit.IsIn(kit.Ext(arg[0]), "zip", "gz") {
		return
	} else if kit.Contains(arg[0], "/dist/", "/bin/", "/log/") {
		return
	} else if strings.HasPrefix(arg[0], "usr/volcanos/publish/") && !strings.HasSuffix(arg[0], "/proto.js") {
		return
	}
	switch arg[0] {
	case ice.SRC_VERSION_GO, ice.SRC_BINPACK_GO:
		return
	case ice.ETC_LOCAL_SHY:
		fmt.Fprintf(w, "        \"%s\": \"%s\",\n", kit.Select(arg[0], arg, 1), "")
		return
	}
	if f, e := nfs.OpenFile(m, arg[0]); !m.Warn(e, ice.ErrNotFound, arg[0]) {
		defer f.Close()
		if b, e := ioutil.ReadAll(f); !m.Warn(e, ice.ErrNotValid, arg[0]) {
			kit.If(len(b) > 1<<20, func() { m.Warn("too large %s %s", arg[0], len(b)) })
			fmt.Fprintf(w, "        \"%s\": \"%s\",\n", kit.Select(arg[0], arg, 1), base64.StdEncoding.EncodeToString(b))
		}
	}
}
func _binpack_dir(m *ice.Message, w io.Writer, dir string) {
	nfs.DirDeepAll(m, dir, nfs.PWD, func(value ice.Maps) { _binpack_file(m, w, path.Join(dir, value[nfs.PATH])) })
}
func _binpack_all(m *ice.Message) {
	w, p, e := nfs.CreateFile(m, ice.SRC_BINPACK_GO)
	m.Assert(e)
	defer w.Close()
	defer m.Echo(p)
	fmt.Fprintln(w, nfs.Template(m, ice.SRC_BINPACK_GO))
	defer fmt.Fprintln(w, nfs.Template(m, "binpack_end.go"))
	defer fmt.Fprint(w, lex.TB)
	nfs.OptionFiles(m, nfs.DiskFile)
	kit.For([]string{ice.USR_VOLCANOS, ice.USR_INTSHELL, ice.SRC}, func(p string) { _binpack_dir(m, w, p) })
	kit.For([]string{
		ice.ETC_MISS_SH, ice.ETC_INIT_SHY, ice.ETC_LOCAL_SHY, ice.ETC_EXIT_SHY, ice.ETC_PATH,
		ice.README_MD, ice.MAKEFILE, ice.LICENSE, ice.GO_MOD, ice.GO_SUM,
	}, func(p string) { _binpack_file(m, w, p) })
	list, cache := map[string]string{}, GoCache(m)
	for k := range ice.Info.File {
		switch ls := kit.Split(k, nfs.PS); ls[1] {
		case ice.SRC:
		case ice.USR:
			list[path.Join(kit.Slice(ls, 1, -1)...)] = ""
		default:
			p := path.Join(cache, path.Join(kit.Slice(ls, 1, -1)...))
			list[path.Join(nfs.USR, strings.Split(ls[3], mdb.AT)[0], path.Join(kit.Slice(ls, 4)...))] = p
		}
	}
	for _, k := range kit.SortedKey(list) {
		v := kit.Select(k, list[k])
		m.Cmd(nfs.DIR, nfs.PWD, nfs.PATH, kit.Dict(nfs.DIR_ROOT, v, nfs.DIR_REG, kit.ExtReg(kit.Split(mdb.Config(m, lex.EXTREG))...))).Table(func(value ice.Maps) {
			if kit.HasPrefix(k, ice.USR_ICEBERGS) && !nfs.Exists(m, ice.USR_ICEBERGS) || m.Option(ice.MSG_USERPOD) != "" {
				return
			}
			_binpack_file(m, w, kit.Path(v, value[nfs.PATH]), path.Join(k, value[nfs.PATH]))
		})
	}
	mdb.HashSelects(m).Sort(nfs.PATH).Table(func(value ice.Maps) {
		if strings.HasSuffix(value[nfs.PATH], nfs.PS) {
			_binpack_dir(m, w, value[nfs.PATH])
		} else {
			_binpack_file(m, w, value[nfs.PATH])
		}
	})
	kit.If(nfs.Exists(m, ice.USR_RELEASE) && m.Option(ice.MSG_USERPOD) == "", func() {
		m.Option(nfs.DIR_REG, kit.ExtReg(nfs.SHY))
		_binpack_dir(m, w, ice.USR_RELEASE)
	})
}

const BINPACK = "binpack"

func init() {
	Index.MergeCommands(ice.Commands{
		BINPACK: {Name: "binpack path auto create insert", Help: "打包", Actions: ice.MergeActions(ice.Actions{
			mdb.CREATE: {Name: "create path", Hand: func(m *ice.Message, arg ...string) { _binpack_all(m) }},
			mdb.INSERT: {Name: "insert path*", Hand: func(m *ice.Message, arg ...string) { mdb.HashCreate(m) }},
		}, mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path", lex.EXTREG, "sh,shy,py,js,css,html,png,jpg"))},
	})
}
func GoCache(m *ice.Message) string {
	return kit.GetValid(
		func() string { return GoEnv(m, GOMODCACHE) },
		func() string { return kit.Select(kit.HomePath(GO)+nfs.PS, GoEnv(m, GOPATH)) + "/pkg/mod/" },
		func() string { return ice.USR_REQUIRE },
	)
}
