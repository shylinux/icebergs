package nfs

import (
	"io"
	"io/ioutil"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/icebergs/base/mdb"
	kit "shylinux.com/x/toolkits"
)

const PACK = "pack"

func init() {
	pack := PackFile
	Index.MergeCommands(ice.Commands{
		PACK: {Name: "pack path auto upload create", Help: "文件系统", Actions: ice.Actions{
			mdb.UPLOAD: {Name: "upload", Help: "上传", Hand: func(m *ice.Message, arg ...string) {
				if b, h, e := m.R.FormFile(mdb.UPLOAD); m.Assert(e) {
					defer b.Close()
					if f, p, e := pack.CreateFile(path.Join(m.Option(PATH), h.Filename)); m.Assert(e) {
						defer f.Close()
						if n, e := io.Copy(f, b); e == nil {
							m.Log_IMPORT(FILE, p, SIZE, n)
						}
					}
				}
			}},
			mdb.CREATE: {Name: "create path=h1/h2/hi.txt text=hello", Help: "创建", Hand: func(m *ice.Message, arg ...string) {
				if f, _, e := pack.CreateFile(m.Option(PATH)); e == nil {
					defer f.Close()
					f.Write([]byte(m.Option(mdb.TEXT)))
				}
			}},
			mdb.REMOVE: {Name: "remove", Help: "删除", Hand: func(m *ice.Message, arg ...string) {
				pack.Remove(path.Clean(m.Option(PATH)))
			}},
		}, Hand: func(m *ice.Message, arg ...string) {
			p := kit.Select("", arg, 0)
			if p != "" && !strings.HasSuffix(p, PS) {
				if f, e := pack.OpenFile(p); e == nil {
					defer f.Close()
					if b, e := ioutil.ReadAll(f); e == nil {
						m.Echo(string(b))
					}
				}
				return
			}

			ls, _ := pack.ReadDir(p)
			for _, f := range ls {
				m.Push(mdb.TIME, f.ModTime().Format(ice.MOD_TIME))
				m.Push(PATH, path.Join(p, f.Name())+kit.Select("", PS, f.IsDir()))
				m.Push(SIZE, f.Size())
			}
			m.Sort("time,path")
			m.PushAction(mdb.REMOVE)
			m.StatusTimeCount()
		}},
	})
}
