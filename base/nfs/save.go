package nfs

import (
	"io"
	"os"
	"path"
	"strings"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _save_file(m *ice.Message, name string, text ...string) {
	if f, p, e := kit.Create(path.Join(m.Option(DIR_ROOT), name)); m.Assert(e) {
		defer f.Close()

		for _, v := range text {
			if n, e := f.WriteString(v); m.Assert(e) {
				m.Log_EXPORT(FILE, p, kit.MDB_SIZE, n)
			}
		}
		m.Echo(p)
	}
}
func _defs_file(m *ice.Message, name string, text ...string) {
	if _, e := os.Stat(path.Join(m.Option(DIR_ROOT), name)); os.IsNotExist(e) {
		_save_file(m, name, text...)
	}
}
func _push_file(m *ice.Message, name string, text ...string) {
	p := path.Join(m.Option(DIR_ROOT), name)
	if strings.Contains(p, ice.PS) {
		os.MkdirAll(path.Dir(p), ice.MOD_DIR)
	}

	if f, e := os.OpenFile(p, os.O_WRONLY|os.O_APPEND|os.O_CREATE, ice.MOD_FILE); m.Assert(e) {
		defer f.Close()

		for _, k := range text {
			if n, e := f.WriteString(k); m.Assert(e) {
				m.Log_EXPORT(FILE, p, kit.MDB_SIZE, n)
			}
		}
		m.Echo(p)
	}
}
func _copy_file(m *ice.Message, name string, from ...string) {
	if f, p, e := kit.Create(path.Join(m.Option(DIR_ROOT), name)); m.Assert(e) {
		defer f.Close()

		for _, v := range from {
			if s, e := os.Open(v); !m.Warn(e, ice.ErrNotFound, name) {
				defer s.Close()

				if n, e := io.Copy(f, s); !m.Warn(e, ice.ErrNotFound, name) {
					m.Log_IMPORT(FILE, v, kit.MDB_SIZE, n)
					m.Log_EXPORT(FILE, p, kit.MDB_SIZE, n)
				}
			}
		}
		m.Echo(p)
	}
}
func _link_file(m *ice.Message, name string, from string) {
	if from == "" {
		return
	}
	os.Remove(name)
	os.MkdirAll(path.Dir(name), ice.MOD_DIR)
	m.Warn(os.Link(from, name), ice.ErrNotFound, from)
	m.Echo(name)
}

const SAVE = "save"
const DEFS = "defs"
const PUSH = "push"
const COPY = "copy"
const LINK = "link"

func init() {
	Index.Merge(&ice.Context{Commands: map[string]*ice.Command{
		SAVE: {Name: "save file text...", Help: "保存", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 1 {
				arg = append(arg, m.Option(kit.MDB_CONTENT))
			}
			_save_file(m, arg[0], arg[1:]...)
		}},
		DEFS: {Name: "defs file text...", Help: "默认", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_defs_file(m, arg[0], arg[1:]...)
		}},
		PUSH: {Name: "push file text...", Help: "追加", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			if len(arg) == 1 {
				arg = append(arg, m.Option(kit.MDB_CONTENT))
			}
			_push_file(m, arg[0], arg[1:]...)
		}},
		COPY: {Name: "copy file from...", Help: "复制", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			for _, file := range arg[1:] {
				if _, e := os.Stat(file); e == nil {
					_copy_file(m, arg[0], arg[1:]...)
				}
			}
		}},
		LINK: {Name: "link file from", Help: "链接", Hand: func(m *ice.Message, c *ice.Context, cmd string, arg ...string) {
			_link_file(m, arg[0], arg[1])
		}},
	}})
}
