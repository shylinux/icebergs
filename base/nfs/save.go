package nfs

import (
	"fmt"
	"io"
	"path"

	ice "shylinux.com/x/icebergs"
	kit "shylinux.com/x/toolkits"
)

func _defs_file(m *ice.Message, name string, text ...string) {
	if Exists(m, path.Join(m.Option(DIR_ROOT), name)) {
		return
	}
	for i, v := range text {
		if b, e := kit.Render(v, m); !m.Warn(e) {
			text[i] = string(b)
		}
	}
	_save_file(m, name, text...)
}
func _save_file(m *ice.Message, name string, text ...string) {
	Create(m, path.Join(m.Option(DIR_ROOT), name), func(w io.Writer, p string) {
		defer m.Echo(p)
		kit.For(text, func(s string) { Save(m, w, s, func(n int) { m.Logs(SAVE, FILE, p, SIZE, n) }) })
	})
}
func _push_file(m *ice.Message, name string, text ...string) {
	Append(m, path.Join(m.Option(DIR_ROOT), name), func(w io.Writer, p string) {
		defer m.Echo(p)
		kit.For(text, func(s string) { Save(m, w, s, func(n int) { m.Logs(SAVE, FILE, p, SIZE, n) }) })
	})
}
func _copy_file(m *ice.Message, name string, from ...string) {
	Create(m, path.Join(m.Option(DIR_ROOT), name), func(w io.Writer, p string) {
		defer m.Echo(p)
		kit.For(from, func(f string) {
			Open(m, path.Join(m.Option(DIR_ROOT), f), func(r io.Reader) {
				Copy(m, w, r, func(n int) { m.Logs(LOAD, FILE, f, SIZE, n).Logs(SAVE, FILE, p, SIZE, n) })
			})
		})
	})
}
func _link_file(m *ice.Message, name string, from string) {
	if m.Warn(from == "", ice.ErrNotValid, FROM) {
		return
	}
	name = path.Join(m.Option(DIR_ROOT), name)
	from = path.Join(m.Option(DIR_ROOT), from)
	if m.Warn(!Exists(m, from), ice.ErrNotFound, from) {
		return
	}
	Remove(m, name)
	if MkdirAll(m, path.Dir(name)); m.Warn(Link(m, from, name)) && m.Warn(Symlink(m, from, name), ice.ErrWarn, from) {
		return
	}
	m.Logs(SAVE, FILE, name, FROM, from).Echo(name)
}

const (
	CONTENT = "content"
	OFFSET  = "offset"
	ALIAS   = "alias"
	FROM    = "from"
	TO      = "to"
)
const DEFS = "defs"
const SAVE = "save"
const PUSH = "push"
const COPY = "copy"
const LINK = "link"
const LOAD = "load"

func init() {
	Index.MergeCommands(ice.Commands{
		DEFS: {Name: "defs file text run", Help: "默认", Hand: func(m *ice.Message, arg ...string) {
			OptionFiles(m, DiskFile)
			_defs_file(m, arg[0], arg[1:]...)
		}},
		SAVE: {Name: "save file text run", Help: "保存", Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) == 1, func() { arg = append(arg, m.Option(CONTENT)) })
			_save_file(m, arg[0], arg[1:]...)
		}},
		PUSH: {Name: "push file text run", Help: "追加", Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) == 1, func() { arg = append(arg, m.Option(CONTENT)) })
			_push_file(m, arg[0], arg[1:]...)
		}},
		COPY: {Name: "copy file from run", Help: "复制", Hand: func(m *ice.Message, arg ...string) {
			kit.If(len(arg) > 1 && Exists(m, arg[1]), func() { _copy_file(m, arg[0], arg[1:]...) })
		}},
		LINK: {Name: "link file from run", Help: "链接", Hand: func(m *ice.Message, arg ...string) {
			_link_file(m, arg[0], arg[1])
		}},
	})
}
func Create(m *ice.Message, p string, cb ice.Any) {
	if f, p, e := CreateFile(m, p); !m.Warn(e) {
		defer f.Close()
		switch cb := cb.(type) {
		case func(io.Writer, string):
			cb(f, p)
		case func(io.Writer):
			cb(f)
		default:
			m.ErrorNotImplement(cb)
		}
	}
}
func Append(m *ice.Message, p string, cb ice.Any) {
	if f, p, e := AppendFile(m, p); !m.Warn(e) {
		defer f.Close()
		switch cb := cb.(type) {
		case func(io.Writer, string):
			cb(f, p)
		case func(io.Writer):
			cb(f)
		default:
			m.ErrorNotImplement(cb)
		}
	}
}
func Save(m *ice.Message, w io.Writer, s string, cb ice.Any) {
	if n, e := fmt.Fprint(w, s); !m.Warn(e) {
		switch cb := cb.(type) {
		case func(int):
			cb(n)
		default:
			m.ErrorNotImplement(cb)
		}
	}
}
func Copy(m *ice.Message, w io.Writer, r io.Reader, cb ice.Any) {
	if n, e := io.Copy(w, r); !m.Warn(e) {
		switch cb := cb.(type) {
		case func(int):
			cb(int(n))
		default:
			m.ErrorNotImplement(cb)
		}
	}
}
func CopyStream(m *ice.Message, to io.WriteCloser, from io.ReadCloser, cache, total int, cb ice.Any) {
	count, buf := 0, make([]byte, cache)
	for {
		n, e := from.Read(buf)
		to.Write(buf[0:n])
		if count += n; count > total {
			total = count
		}
		switch value := count * 100 / total; cb := cb.(type) {
		case func(int, int, int):
			cb(count, total, value)
		case func(int, int):
			cb(count, total)
		case nil:
		default:
			m.ErrorNotImplement(cb)
		}
		if e == io.EOF || m.Warn(e) {
			break
		}
	}
}
func CopyFile(m *ice.Message, to, from string, cb func([]byte, int) []byte) {
	Open(m, from, func(r io.Reader) {
		Create(m, to, func(w io.Writer) {
			offset, buf := 0, make([]byte, 1024*1024)
			for {
				n, _ := r.Read(buf)
				if n, _ = w.Write(cb(buf[:n], offset)); n == 0 {
					break
				}
				offset += n
			}
			m.Logs(SAVE, FILE, to, FROM, from, SIZE, offset)
		})
	})
}
func Pipe(m *ice.Message, cb func(string)) io.WriteCloser {
	r, w := io.Pipe()
	m.Go(func() { kit.For(r, cb) })
	return w
}
