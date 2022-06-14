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
	Domain   string
	Version  string
	HostName string
	UserName string
}

var Info = struct {
	Make MakeInfo

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

	Help  string
	Route map[string]string // 路由命令
	File  map[string]string // 文件命令
	Pack  map[string][]byte // 打包文件
	Dump  func(w io.Writer, name string, cb func(string)) bool
	Log   func(m *Message, p, l, s string)

	render map[string]func(*Message, string, ...Any) string
	names  Map
}{
	Help: `
^_^      欢迎使用冰山框架       ^_^
^_^  Welcome to Icebergs World  ^_^

report: shylinuxc@gmail.com
server: https://shylinux.com
source: https://shylinux.com/x/icebergs
`,
	Route: map[string]string{},
	File:  map[string]string{},
	Pack:  map[string][]byte{},
	Dump:  func(w io.Writer, name string, cb func(string)) bool { return false },
	Log:   func(m *Message, p, l, s string) {},

	render: map[string]func(*Message, string, ...Any) string{},
	names:  Map{},
}

func FileURI(dir string) string {
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
	if kit.FileExists(path.Join("src", dir)) {
		return path.Join("/require/src/", dir)
	}
	return dir
}
func FileCmd(dir string) string {
	dir = strings.Split(dir, DF)[0]
	dir = strings.ReplaceAll(dir, ".js", ".go")
	dir = strings.ReplaceAll(dir, ".sh", ".go")
	return FileURI(dir)
}
func AddFileCmd(dir, key string) {
	Info.File[FileCmd(dir)] = key
}
func GetFileCmd(dir string) string {
	if strings.HasPrefix(dir, "require/") {
		dir = "/" + dir
	}
	for _, dir := range []string{dir, "/require/" + Info.Make.Module + "/" + dir, "/require/" + Info.Make.Module + "/src/" + dir} {
		if cmd, ok := Info.File[FileCmd(dir)]; ok {
			return cmd
		}
		p := path.Dir(dir)
		if cmd, ok := Info.File[FileCmd(path.Join(p, path.Base(p)+".go"))]; ok {
			return cmd
		}
		for k, v := range Info.File {
			if strings.HasPrefix(k, p) {
				return v
			}
		}
	}
	return ""
}
func FileRequire(n int) string {
	p := kit.Split(kit.FileLine(n, 100), DF)[0]
	if strings.Contains(p, "go/pkg/mod") {
		return path.Join("/require", strings.Split(p, "go/pkg/mod")[1])
	}
	return path.Join("/require/"+kit.ModPath(n), path.Base(p))
}
