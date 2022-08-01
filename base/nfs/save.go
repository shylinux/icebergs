package nfs

import (
	"fmt"
	"io"
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

		for _, v := range text {
			if n, e := fmt.Fprint(f, v); m.Assert(e) {
				m.Log_EXPORT(FILE, p, SIZE, n)
			}
		}
		m.Echo(p)
	}
}
func _push_file(m *ice.Message, name string, text ...string) {
	if f, p, e := AppendFile(m, path.Join(m.Option(DIR_ROOT), name)); m.Assert(e) {
		defer f.Close()

		for _, k := range text {
			if n, e := fmt.Fprint(f, k); m.Assert(e) {
				m.Log_EXPORT(FILE, p, SIZE, n)
			}
		}
		m.Echo(p)
	}
}
func _copy_file(m *ice.Message, name string, from ...string) {
	if f, p, e := CreateFile(m, path.Join(m.Option(DIR_ROOT), name)); m.Assert(e) {
		defer f.Close()

		for _, v := range from {
			if s, e := OpenFile(m, path.Join(m.Option(DIR_ROOT), v)); !m.Warn(e, ice.ErrNotFound, name) {
				defer s.Close()

				if n, e := io.Copy(f, s); !m.Warn(e, ice.ErrNotFound, name) {
					m.Log_IMPORT(FILE, v, SIZE, n)
					m.Log_EXPORT(FILE, p, SIZE, n)
				}
			}
		}
		m.Echo(p)
	}
}
func _link_file(m *ice.Message, name string, from string) {
	if m.Warn(from == "", ice.ErrNotValid, from) {
		return
	}
	name = path.Join(m.Option(DIR_ROOT), name)
	from = path.Join(m.Option(DIR_ROOT), from)
	if m.Warn(!ExistsFile(m, from), ice.ErrNotFound, from) {
		return
	}
	Remove(m, name)
	MkdirAll(m, path.Dir(name))
	if m.Warn(Link(m, from, name)) {
		if m.Warn(Symlink(m, from, name), ice.ErrWarn, from) {
			return
		}
	}
	m.Log_EXPORT(FILE, name, FROM, from)
	m.Echo(name)
}

const (
	CONTENT = "content"
	FROM    = "from"
)
const DEFS = "defs"
const SAVE = "save"
const PUSH = "push"
const COPY = "copy"
const LINK = "link"

func init() {
	Index.MergeCommands(ice.Commands{
		DEFS: {Name: "defs file text...", Help: "默认", Hand: func(m *ice.Message, arg ...string) {
			_defs_file(m, arg[0], arg[1:]...)
		}},
		SAVE: {Name: "save file text...", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 1 {
				arg = append(arg, m.Option(CONTENT))
			}
			_save_file(m, arg[0], arg[1:]...)
		}},
		PUSH: {Name: "push file text...", Help: "追加", Hand: func(m *ice.Message, arg ...string) {
			if len(arg) == 1 {
				arg = append(arg, m.Option(CONTENT))
			}
			_push_file(m, arg[0], arg[1:]...)
		}},
		COPY: {Name: "copy file from...", Help: "复制", Hand: func(m *ice.Message, arg ...string) {
			for _, file := range arg[1:] {
				if kit.FileExists(file) {
					_copy_file(m, arg[0], arg[1:]...)
					return
				}
			}
		}},
		LINK: {Name: "link file from", Help: "链接", Hand: func(m *ice.Message, arg ...string) {
			_link_file(m, arg[0], arg[1])
		}},
	})
}
