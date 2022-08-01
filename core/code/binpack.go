package code

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	"shylinux.com/x/icebergs/base/nfs"
	kit "shylinux.com/x/toolkits"
)

func _binpack_file(m *ice.Message, arg ...string) string { // file name
	if f, e := os.Open(arg[0]); e == nil {
		defer f.Close()
		if b, e := ioutil.ReadAll(f); e == nil && len(b) > 0 {
			return fmt.Sprintf("        \"%s\": \"%s\",", kit.Select(arg[0], arg, 1), base64.StdEncoding.EncodeToString(b))
		}
	}
	return fmt.Sprintf("        // \"%s\": \"%s\",", kit.Select(arg[0], arg, 1), "")
}
func _binpack_dir(m *ice.Message, f *os.File, dir string) {
	m.Option(nfs.DIR_ROOT, dir)
	m.Option(nfs.DIR_DEEP, true)
	m.Option(nfs.DIR_TYPE, nfs.CAT)

	m.Cmd(nfs.DIR, nfs.PWD).Sort(nfs.PATH).Tables(func(value ice.Maps) {
		switch path.Base(value[nfs.PATH]) {
		case "go.mod", "go.sum", "binpack.go", "version.go":
			return
		}
		switch strings.Split(value[nfs.PATH], ice.PS)[0] {
		case "var", "polaris", "website":
			return
		}
		fmt.Fprintln(f, _binpack_file(m, path.Join(dir, value[nfs.PATH])))
	})
	fmt.Fprintln(f)
}

func _binpack_can(m *ice.Message, f *os.File, dir string) {
	m.Option(nfs.DIR_ROOT, dir)
	m.Option(nfs.DIR_DEEP, true)
	m.Option(nfs.DIR_TYPE, nfs.CAT)

	for _, k := range []string{ice.FAVICON, ice.PROTO_JS, ice.FRAME_JS} {
		// fmt.Fprintln(f, _binpack_file(m, path.Join(dir, k), ice.PS+k))
		fmt.Fprintln(f, _binpack_file(m, path.Join(dir, k), path.Join(ice.USR_VOLCANOS, k)))
	}
	for _, k := range []string{LIB, PAGE, PANEL, PLUGIN, "publish/client/nodejs/"} {
		m.Cmd(nfs.DIR, k).Sort(nfs.PATH).Tables(func(value ice.Maps) {
			// fmt.Fprintln(f, _binpack_file(m, path.Join(dir, value[nfs.PATH]), ice.PS+value[nfs.PATH]))
			fmt.Fprintln(f, _binpack_file(m, path.Join(dir, value[nfs.PATH]), path.Join(ice.USR_VOLCANOS, value[nfs.PATH])))
		})
	}
	fmt.Fprintln(f)
}
func _binpack_ctx(m *ice.Message, f *os.File) {
	_binpack_dir(m, f, ice.SRC)
}

const BINPACK = "binpack"

func init() {
	Index.MergeCommands(ice.Commands{
		BINPACK: {Name: "binpack path auto create remove export", Help: "打包", Actions: ice.MergeAction(ice.Actions{
			ice.CTX_INIT: {Hand: func(m *ice.Message, arg ...string) {
				if kit.FileExists(path.Join(ice.USR_VOLCANOS, ice.PROTO_JS)) {
					m.Cmd(BINPACK, mdb.REMOVE)
					return
				}
			}},
			mdb.CREATE: {Name: "create", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if f, p, e := kit.Create(ice.SRC_BINPACK_GO); m.Assert(e) {
					defer f.Close()
					defer m.Echo(p)

					fmt.Fprintln(f, `package main

import (
	"encoding/base64"
	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/nfs"
)

func init() {
`)
					defer fmt.Fprintln(f, `}`)

					defer fmt.Fprintln(f, `
	for k, v := range pack {
		if b, e := base64.StdEncoding.DecodeString(v); e == nil {
			nfs.PackFile.WriteFile(k, b)
		}
	}
`)

					fmt.Fprintln(f, `	pack := ice.Maps{`)
					defer fmt.Fprintln(f, `	}`)

					if kit.FileExists(ice.USR_VOLCANOS) && kit.FileExists(ice.USR_INTSHELL) && m.Option(ice.MSG_USERPOD) == "" {
						_binpack_can(m, f, ice.USR_VOLCANOS)
						_binpack_dir(m, f, ice.USR_INTSHELL)
					}
					_binpack_ctx(m, f)

					fmt.Fprintln(f, _binpack_file(m, ice.ETC_MISS_SH))
					fmt.Fprintln(f, _binpack_file(m, ice.ETC_INIT_SHY))
					fmt.Fprintln(f, _binpack_file(m, ice.ETC_EXIT_SHY))
					fmt.Fprintln(f)

					fmt.Fprintln(f, _binpack_file(m, ice.LICENSE))
					fmt.Fprintln(f, _binpack_file(m, ice.MAKEFILE))
					fmt.Fprintln(f, _binpack_file(m, ice.README_MD))
					fmt.Fprintln(f)

					m.Cmd(mdb.SELECT, m.PrefixKey(), "", mdb.HASH, ice.OptionFields(nfs.PATH)).Tables(func(value ice.Maps) {
						if s, e := os.Stat(value[nfs.PATH]); e == nil {
							if s.IsDir() {
								_binpack_dir(m, f, value[nfs.PATH])
							} else {
								fmt.Fprintln(f, _binpack_file(m, value[nfs.PATH]))
							}
						}
					})
				}
			}},
			mdb.INSERT: {Name: "insert", Help: "添加", Hand: func(m *ice.Message, arg ...string) {
				m.Cmd(mdb.INSERT, m.PrefixKey(), "", mdb.HASH, nfs.PATH, arg[0])
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				ice.Info.Pack = map[string][]byte{}
			}},
			mdb.EXPORT: {Name: "export", Help: "导出", Hand: func(m *ice.Message, arg ...string) {
				for key, value := range ice.Info.Pack {
					if strings.HasPrefix(key, ice.PS) {
						key = ice.USR_VOLCANOS + key
					}
					m.Log_EXPORT(nfs.FILE, kit.WriteFile(key, value), nfs.SIZE, len(value))
				}
			}},
		}, mdb.HashAction(mdb.SHORT, nfs.PATH)), Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 0 {
				for k, v := range ice.Info.Pack {
					m.Push(nfs.PATH, k).Push(nfs.SIZE, len(v))
				}
				m.Sort(nfs.PATH)
				return
			}
			m.Echo(string(ice.Info.Pack[arg[0]]))
		}},
	})
}
