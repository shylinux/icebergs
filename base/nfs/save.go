package nfs

import (
	"fmt"
	"io"
	"os"
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _defs_file(m *ice.Message, name string, text ...string) {
	if ExistsFile(m, path.Join(m.Option(DIR_ROOT), name)) {
		return
	}
	for i, v := range text {
		b, _ := kit.Render(v, m)
		text[i] = string(b)
	}
	_save_file(m, name, text...)
}
func _save_file(m *ice.Message, name string, text ...string) {
	if f, p, e := CreateFile(m, path.Join(m.Option(DIR_ROOT), name)); m.Assert(e) {
		defer f.Close()
		defer m.Echo(p)
		for _, v := range text {
			if n, e := fmt.Fprint(f, v); m.Assert(e) {
				m.Logs(SAVE, FILE, p, SIZE, n)
			}
		}
	}
}
func _push_file(m *ice.Message, name string, text ...string) {
	if f, p, e := AppendFile(m, path.Join(m.Option(DIR_ROOT), name)); m.Assert(e) {
		defer f.Close()
		defer m.Echo(p)
		for _, k := range text {
			if n, e := fmt.Fprint(f, k); m.Assert(e) {
				m.Logs(SAVE, FILE, p, SIZE, n)
			}
		}
	}
}
func _copy_file(m *ice.Message, name string, from ...string) {
	if f, p, e := CreateFile(m, path.Join(m.Option(DIR_ROOT), name)); m.Assert(e) {
		defer f.Close()
		defer m.Echo(p)
		for _, v := range from {
			if s, e := OpenFile(m, path.Join(m.Option(DIR_ROOT), v)); !m.Warn(e, ice.ErrNotFound, v) {
				defer s.Close()
				if n, e := io.Copy(f, s); !m.Warn(e, ice.ErrNotValid, v) {
					m.Logs(LOAD, FILE, v, SIZE, n)
					m.Logs(SAVE, FILE, p, SIZE, n)
				}
			}
		}
	}
}
func _link_file(m *ice.Message, name string, from string) {
	if m.Warn(from == "", ice.ErrNotValid, FROM) {
		return
	}
	name = path.Join(m.Option(DIR_ROOT), name)
	from = path.Join(m.Option(DIR_ROOT), from)
	if m.Warn(!ExistsFile(m, from), ice.ErrNotFound, from) {
		return
	}
	Remove(m, name)
	if MkdirAll(m, path.Dir(name)); m.Warn(Link(m, from, name)) && m.Warn(Symlink(m, from, name), ice.ErrWarn, from) {
		return
	}
	m.Logs(SAVE, FILE, name, FROM, from)
	m.Echo(name)
}

const (
	CONTENT = "content"
	ALIAS   = "alias"
	FROM    = "from"
	TO      = "to"
)
const LOAD = "load"
const DEFS = "defs"
const SAVE = "save"
const PUSH = "push"
const COPY = "copy"
const LINK = "link"

func init() {
	Index.MergeCommands(ice.Commands{
		DEFS: {Name: "defs file text run", Help: "默认", Hand: func(m *ice.Message, arg ...string) {
			OptionFiles(m, DiskFile)
			_defs_file(m, arg[0], arg[1:]...)
		}},
		SAVE: {Name: "save file text run", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 1 {
				arg = append(arg, m.Option(CONTENT))
			}
			_save_file(m, arg[0], arg[1:]...)
		}},
		PUSH: {Name: "push file text run", Help: "追加", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 1 {
				arg = append(arg, m.Option(CONTENT))
			}
			_push_file(m, arg[0], arg[1:]...)
		}},
		COPY: {Name: "copy file from run", Help: "复制", Hand: func(m *ice.Message, arg ...string) {
			for _, file := range arg[1:] {
				if ExistsFile(m, file) {
					_copy_file(m, arg[0], arg[1:]...)
					return
				}
			}
		}},
		LINK: {Name: "link file from run", Help: "链接", Hand: func(m *ice.Message, arg ...string) {
			_link_file(m, arg[0], arg[1])
		}},
	})
}
func Copy(m *ice.Message, cb func([]byte, int) []byte, to, from string) {
	if _from, e := os.Open(from); m.Assert(e) {
		defer _from.Close()
		if _to, e := os.Create(to); m.Assert(e) {
			defer _to.Close()
			offset, buf := 0, make([]byte, 1024*1024)
			for {
				n, _ := _from.Read(buf)
				if n, _ = _to.Write(cb(buf[:n], offset)); n == 0 {
					break
				}
				offset += n
			}
			m.Logs(SAVE, FILE, to, FROM, from, SIZE, offset)
		}
	}
}
