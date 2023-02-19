package code

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/ctx"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _binpack_file(m *ice.Message, w io.Writer, arg ...string) {
	if strings.HasPrefix(arg[0], "usr/volcanos/publish/") && !strings.HasSuffix(arg[0], "/proto.js") {
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
	nfs.DirDeepAll(m, dir, nfs.PWD, func(value ice.Maps) {
		_binpack_file(m, w, path.Join(dir, value[nfs.PATH]))
	})
}
func _binpack_all(m *ice.Message) {
	nfs.OptionFiles(m, nfs.DiskFile)
	if w, p, e := nfs.CreateFile(m, ice.SRC_BINPACK_GO); m.Assert(e) {
		defer w.Close()
		defer m.Echo(p)
		fmt.Fprintln(w, _binpack_template)
		defer fmt.Fprintln(w, _binpack_template_end)
		_binpack_dir(m, w, ice.USR_VOLCANOS)
		_binpack_dir(m, w, ice.USR_INTSHELL)
		_binpack_dir(m, w, ice.SRC)
		_binpack_file(m, w, ice.ETC_MISS_SH)
		_binpack_file(m, w, ice.ETC_INIT_SHY)
		_binpack_file(m, w, ice.ETC_EXIT_SHY)
		_binpack_file(m, w, ice.README_MD)
		_binpack_file(m, w, ice.MAKEFILE)
		_binpack_file(m, w, ice.LICENSE)
		list := map[string]bool{}
		ctx.TravelCmd(m, func(key, file, line string) {
			dir := path.Dir(file)
			if strings.HasPrefix(dir, ice.SRC) {
				return
			}
			if list[dir] {
				return
			}
			list[dir] = true
			m.Cmd(nfs.DIR, dir, nfs.PATH, kit.Dict(nfs.DIR_ROOT, nfs.PWD, nfs.DIR_REG, kit.ExtReg("(sh|shy|js)"))).Tables(func(value ice.Maps) {
				if list[value[nfs.PATH]] {
					return
				}
				if list[value[nfs.PATH]] = true; strings.Contains(value[nfs.PATH], "/go/pkg/mod/") {
					value[nfs.PATH] = "/require/" + strings.Split(value[nfs.PATH], "/go/pkg/mod/")[1]
				}
				_binpack_file(m, w, value[nfs.PATH])
			})
		})
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
			mdb.CREATE: {Name: "create", Hand: func(m *ice.Message, arg ...string) { _binpack_all(m) }},
			mdb.INSERT: {Name: "insert path*", Hand: func(m *ice.Message, arg ...string) { mdb.HashCreate(m) }},
		}, mdb.HashAction(mdb.SHORT, nfs.PATH, mdb.FIELD, "time,path"))},
	})
}

var _binpack_template = `package main

import (
	"encoding/base64"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
)

func init() {
	pack := ice.Maps{`
var _binpack_template_end = `	}
	nfs.PackFile.RemoveAll(ice.SRC)
	for k, v := range pack {
		if b, e := base64.StdEncoding.DecodeString(v); e == nil {
			nfs.PackFile.WriteFile(k, b)
		}
	}
}`
