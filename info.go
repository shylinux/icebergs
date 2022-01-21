package ice

import (
	"io"
	"strings"

	kit "shylinux.com/x/toolkits"
)

var Info = struct {
	HostName string
	PathName string
	UserName string
	PassWord string

	NodeType string
	NodeName string
	CtxShare string
	CtxRiver string

	Make struct {
		Time     string
		Hash     string
		Remote   string
		Branch   string
		Version  string
		HostName string
		UserName string
	}

	Help   string
	Pack   map[string][]byte
	names  map[string]interface{}
	render map[string]func(*Message, string, ...interface{}) string
	Log    func(m *Message, p, l, s string)
	File   map[string]string
}{
	Help: `
^_^      欢迎使用冰山框架       ^_^
^_^  Welcome to Icebergs World  ^_^

report: shylinuxc@gmail.com
server: https://shylinux.com
source: https://shylinux.com/x/icebergs
`,
	Pack:   map[string][]byte{},
	names:  map[string]interface{}{},
	render: map[string]func(*Message, string, ...interface{}) string{},
	File:   map[string]string{},
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
