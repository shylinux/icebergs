package ice

import (
	"io"
	"path"
	"strings"

	kit "shylinux.com/x/toolkits"
)

type MakeInfo struct {
	Path     string
	Time     string
	Hash     string
	Module   string
	Remote   string
	Branch   string
	Version  string
	HostName string
	UserName string
}

var Info = struct {
	HostName string
	PathName string
	UserName string
	PassWord string

	Colors   bool
	Domain   string
	NodeType string
	NodeName string
	CtxShare string
	CtxRiver string

	Make MakeInfo

	Help  string
	Pack  map[string][]byte
	File  map[string]string
	Route map[string]string
	Log   func(m *Message, p, l, s string)

	render map[string]func(*Message, string, ...interface{}) string
	names  map[string]interface{}
}{
	Help: `
^_^      欢迎使用冰山框架       ^_^
^_^  Welcome to Icebergs World  ^_^

report: shylinuxc@gmail.com
server: https://shylinux.com
source: https://shylinux.com/x/icebergs
`,
	Pack:  map[string][]byte{},
	File:  map[string]string{},
	Route: map[string]string{},

	render: map[string]func(*Message, string, ...interface{}) string{},
	names:  map[string]interface{}{},
}

func fileKey(dir string) string {
	dir = strings.Split(dir, DF)[0]
	dir = strings.ReplaceAll(dir, ".js", ".go")
	dir = strings.ReplaceAll(dir, ".sh", ".go")

	if strings.Contains(dir, "go/pkg/mod") {
		return path.Join("/require", strings.Split(dir, "go/pkg/mod")[1])
	}
	if Info.Make.Path != "" && strings.HasPrefix(dir, Info.Make.Path+PS) {
		dir = strings.TrimPrefix(dir, Info.Make.Path+PS)
	}
	if strings.HasPrefix(dir, kit.Path("")+PS) {
		dir = strings.TrimPrefix(dir, kit.Path("")+PS)
	}
	if strings.HasPrefix(dir, SRC) {
		return path.Join("/require", dir)
	}
	if strings.HasPrefix(dir, USR) {
		return path.Join("/require", dir)
	}
	return dir
}
func AddFileKey(dir, key string) {
	Info.File[fileKey(dir)] = key
}
func GetFileKey(dir string) string {
	return Info.File[fileKey(dir)]
}
func Dump(w io.Writer, name string, cb func(string)) bool {
	for _, key := range []string{name, strings.TrimPrefix(name, USR_VOLCANOS)} {
		if b, ok := Info.Pack[key]; ok {
			if cb != nil {
				cb(name)
			}
			w.Write(b)
			return true
		}
	}
	return false
}
func name(name string, value interface{}) string {
	if s, ok := Info.names[name]; ok {
		last := ""
		switch s := s.(type) {
		case *Context:
			last = s.Name
		}
		panic(kit.Format("%s %s %v", ErrExists, name, last))
	}

	Info.names[name] = value
	return name
}
