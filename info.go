package ice

type MakeInfo struct {
	Time     string
	Path     string
	Hash     string
	Domain   string
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

	Domain   string
	NodeType string
	NodeName string
	CtxShare string
	CtxRiver string
	PidPath  string
	Colors   bool

	Help  string
	File  Maps
	Route Maps
	index Map

	merges []MergeHandler
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
	File:  Maps{},
	Route: Maps{},
	index: Map{},

	render: map[string]func(*Message, string, ...Any) string{},
	Save:   func(m *Message, key ...string) *Message { return m },
	Load:   func(m *Message, key ...string) *Message { return m },
	Log:    func(m *Message, p, l, s string) {},
}

type MergeHandler func(*Context, string, *Command, string, *Action) (Handler, Handler)

func AddMerges(h ...MergeHandler) { Info.merges = append(Info.merges, h...) }
