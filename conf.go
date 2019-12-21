package ice

const (
	ICE_CHAN = 10
	ICE_INIT = "_init"
	ICE_EXIT = "_exit"
	ICE_DATE = "2006-01-02"
	ICE_TIME = "2006-01-02 15:04:05"
)
const (
	CTX_STATUS = "status"
	CTX_STREAM = "stream"
	CTX_CONFIG = "config"
)
const (
	MSG_DETAIL = "detail"
	MSG_OPTION = "option"
	MSG_APPEND = "append"
	MSG_RESULT = "result"
)
const (
	MDB_META = "meta"
	MDB_LIST = "list"
	MDB_HASH = "hash"

	MDB_TYPE = "_type"
)
const (
	WEB_PORT = ":9020"
	WEB_SESS = "sessid"
	WEB_TMPL = "render"

	WEB_LOGIN = "_login"
	WEB_SPIDE = "spide"
	WEB_SERVE = "serve"
	WEB_SPACE = "space"
	WEB_STORY = "story"
	WEB_CACHE = "cache"
	WEB_ROUTE = "route"
	WEB_PROXY = "proxy"
)
const (
	GDB_SIGNAL = "signal"
	GDB_TIMER  = "timer"
	GDB_EVENT  = "event"
)
const (
	LOG_CMD   = "cmd"
	LOG_INFO  = "info"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"

	LOG_BEGIN = "begin"
	LOG_START = "start"
	LOG_BENCH = "bench"
	LOG_CLOSE = "close"
)

var Alias = map[string]string{
	CTX_CONFIG: "ctx.config",

	GDB_SIGNAL: "gdb.signal",
	GDB_TIMER:  "gdb.timer",
	GDB_EVENT:  "gdb.event",

	WEB_SPIDE: "web.spide",
	WEB_SERVE: "web.serve",
	WEB_SPACE: "web.space",
	WEB_STORY: "web.story",
	WEB_CACHE: "web.cache",
	WEB_ROUTE: "web.route",
	WEB_PROXY: "web.proxy",

	"note": "web.wiki.note",
}
