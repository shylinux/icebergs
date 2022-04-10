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
	Route: map[string]string{},
	File:  map[string]string{},
	Pack:  map[string][]byte{},
	Dump:  func(w io.Writer, name string, cb func(string)) bool { return false },
	Log:   func(m *Message, p, l, s string) {},

	render: map[string]func(*Message, string, ...interface{}) string{},
	names:  map[string]interface{}{},
}

func FileCmd(dir string) string {
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
func AddFileCmd(dir, key string)   { Info.File[FileCmd(dir)] = key }
func GetFileCmd(dir string) string { return Info.File[FileCmd(dir)] }
