package ice

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
	PidPath  string

	Help  string
	Route Maps // 路由命令
	File  Maps // 文件命令
	names Map

	render map[string]func(*Message, string, ...Any) string
	Save   func(m *Message, key ...string) *Message
	Load   func(m *Message, key ...string) *Message
	Log    func(m *Message, p, l, s string)
}{
	Help: `
^_^      欢迎使用冰山框架       ^_^
^_^  Welcome to Icebergs World  ^_^

report: shylinuxc@gmail.com
server: https://shylinux.com
source: https://shylinux.com/x/icebergs
`,
	Route: Maps{},
	File:  Maps{},
	names: Map{},

	render: map[string]func(*Message, string, ...Any) string{},
	Save:   func(m *Message, key ...string) *Message { return m },
	Load:   func(m *Message, key ...string) *Message { return m },
	Log:    func(m *Message, p, l, s string) {},
}
