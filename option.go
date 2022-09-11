package ice

import (
	"path"
	"strings"
	"time"

	kit "shylinux.com/x/toolkits"
)

type Option struct {
	Name  string
	Value Any
}

func OptionFields(arg ...string) Option { return Option{MSG_FIELDS, kit.Join(arg)} }
func (m *Message) OptionFields(arg ...string) string {
	if len(arg) > 0 {
		m.Option(MSG_FIELDS, kit.Join(arg))
	}
	return kit.Join(kit.Simple(m.Optionv(MSG_FIELDS)))
}
func (m *Message) OptionDefault(arg ...string) string {
	for i := 0; i < len(arg); i += 2 {
		if m.Option(arg[i]) == "" {
			m.Option(arg[i], arg[i+1])
		}
	}
	return m.Option(arg[0])
}
func (m *Message) OptionSimple(key ...string) (res []string) {
	for _, k := range kit.Split(kit.Join(key)) {
		if k == "" || m.Option(k) == "" {
			continue
		}
		res = append(res, k, m.Option(k))
	}
	return
}
func (m *Message) OptionSplit(key ...string) (res []string) {
	for _, k := range kit.Split(kit.Join(key)) {
		res = append(res, m.Option(k))
	}
	return res
}
func (m *Message) OptionCB(key string, cb ...Any) Any {
	if len(cb) > 0 {
		return m.Optionv(kit.Keycb(kit.Select(m.CommandKey(), key)), cb...)
	}
	return m.Optionv(kit.Keycb(kit.Select(m.CommandKey(), key)))
}

func (m *Message) FieldsIsDetail() bool {
	if len(m.meta[MSG_APPEND]) == 2 && m.meta[MSG_APPEND][0] == KEY && m.meta[MSG_APPEND][1] == VALUE {
		return true
	}
	if m.OptionFields() == FIELDS_DETAIL {
		return true
	}
	return false
}
func (m *Message) Fields(length int, fields ...string) string {
	return m.Option(MSG_FIELDS, kit.Select(kit.Select(FIELDS_DETAIL, fields, length), m.Option(MSG_FIELDS)))
}
func (m *Message) Action(arg ...Any) *Message {
	for i, v := range arg {
		switch v.(type) {
		case string:
		default:
			arg[i] = kit.Format(v)
		}
	}
	m.Option(MSG_ACTION, kit.Format(arg))
	return m
}
func (m *Message) Status(arg ...Any) {
	list := kit.List()
	args := kit.Simple(arg)
	for i := 0; i < len(args)-1; i += 2 {
		list = append(list, kit.Dict(NAME, args[i], VALUE, args[i+1]))
	}
	m.Option(MSG_STATUS, kit.Format(list))
}
func (m *Message) StatusTime(arg ...Any) *Message {
	m.Status(TIME, m.Time(), arg, kit.MDB_COST, m.FormatCost())
	return m
}
func (m *Message) StatusTimeCount(arg ...Any) *Message {
	m.Status(TIME, m.Time(), kit.MDB_COUNT, kit.Split(m.FormatSize())[0], arg, kit.MDB_COST, m.FormatCost())
	return m
}
func (m *Message) StatusTimeCountTotal(arg ...Any) {
	m.Status(TIME, m.Time(), kit.MDB_COUNT, kit.Split(m.FormatSize())[0], kit.MDB_TOTAL, arg, kit.MDB_COST, m.FormatCost())
}

func (m *Message) Process(action string, arg ...Any) {
	m.Option(MSG_PROCESS, action)
	m.Option(PROCESS_ARG, arg...)
}
func (m *Message) ProcessLocation(arg ...Any) {
	m.Process(PROCESS_LOCATION, arg...)
}
func (m *Message) ProcessReplace(url string, arg ...Any) {
	m.Process(PROCESS_REPLACE, kit.MergeURL(url, arg...))
}
func (m *Message) ProcessHistory(arg ...Any) {
	m.Process(PROCESS_HISTORY, arg...)
}
func (m *Message) ProcessConfirm(arg ...Any) {
	m.Process(PROCESS_CONFIRM, arg...)
}
func (m *Message) ProcessRefresh(delay string) {
	if d, e := time.ParseDuration(delay); e == nil {
		m.Option("_delay", int(d/time.Millisecond))
	}
	m.Process(PROCESS_REFRESH)
}
func (m *Message) ProcessRefresh3ms()   { m.ProcessRefresh("3ms") }
func (m *Message) ProcessRefresh30ms()  { m.ProcessRefresh("30ms") }
func (m *Message) ProcessRefresh300ms() { m.ProcessRefresh("300ms") }
func (m *Message) ProcessRefresh3s()    { m.ProcessRefresh("3s") }
func (m *Message) ProcessRewrite(arg ...Any) {
	m.Process(PROCESS_REWRITE, arg...)
}
func (m *Message) ProcessDisplay(arg ...Any) {
	m.Process(PROCESS_DISPLAY)
	m.Option(MSG_DISPLAY, arg...)
}

func (m *Message) ProcessStory(arg ...Any) {
	m.Option(MSG_PROCESS, "_story")
	m.Option(PROCESS_ARG, arg...)
}
func (m *Message) ProcessField(arg ...Any) {
	m.Process(PROCESS_FIELD)
	m.Option(FIELD_PREFIX, arg...)
}
func (m *Message) ProcessInner()           { m.Process(PROCESS_INNER) }
func (m *Message) ProcessAgain()           { m.Process(PROCESS_AGAIN) }
func (m *Message) ProcessHold(text ...Any) { m.Process(PROCESS_HOLD, text...) }
func (m *Message) ProcessBack()            { m.Process(PROCESS_BACK) }
func (m *Message) ProcessRich(arg ...Any)  { m.Process(PROCESS_RICH, arg...) }
func (m *Message) ProcessGrow(arg ...Any)  { m.Process(PROCESS_GROW, arg...) }
func (m *Message) ProcessOpen(url string)  { m.Process(PROCESS_OPEN, url) }

func (m *Message) Display(file string, arg ...Any) *Message { // repos local file
	m.Option(MSG_DISPLAY, kit.MergeURL(displayRequire(2, file)[DISPLAY], arg...))
	return m
}
func DisplayLocal(file string, arg ...string) Maps { // /plugin/local/file
	if file == "" {
		file = path.Join(kit.PathName(2), kit.Keys(kit.FileName(2), JS))
	}
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PLUGIN_LOCAL, file)
	}
	return DisplayBase(file, arg...)
}
func DisplayStory(file string, arg ...string) Maps { // /plugin/story/file
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PLUGIN_STORY, file)
	}
	return DisplayBase(file, arg...)
}
func DisplayBase(file string, arg ...string) Maps {
	return Maps{DISPLAY: file, STYLE: kit.Join(arg, SP)}
}
func Display(file string, arg ...string) Maps { // repos local file
	return displayRequire(2, file, arg...)
}
func displayRequire(n int, file string, arg ...string) Maps {
	if file == "" {
		file = kit.Keys(kit.FileName(n+1), JS)
	}
	if !strings.HasPrefix(file, HTTP) && !strings.HasPrefix(file, PS) {
		file = path.Join(PS, path.Join(path.Dir(FileRequire(n+2)), file))
	}
	return DisplayBase(file, arg...)
}
func FileRequire(n int) string {
	p := kit.Split(kit.FileLine(n, 100), DF)[0]
	if strings.Contains(p, "go/pkg/mod") {
		return path.Join("/require", strings.Split(p, "go/pkg/mod")[1])
	}
	return path.Join("/require/"+kit.ModPath(n), path.Base(p))
}
