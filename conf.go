package ice

const (
	ICE_CHAN = 10
	ICE_INIT = "_init"
	ICE_EXIT = "_exit"
	ICE_DATE = "2006-01-02"
	ICE_TIME = "2006-01-02 15:04:05"
)
const (
	CTX_STATUS  = "status"
	CTX_STREAM  = "stream"
	CTX_CONFIG  = "config"
	CTX_COMMAND = "command"
	CTX_CONTEXT = "context"
)
const (
	CLI_RUNTIME = "runtime"
	CLI_SYSTEM  = "system"
)
const (
	MSG_DETAIL = "detail"
	MSG_OPTION = "option"
	MSG_APPEND = "append"
	MSG_RESULT = "result"

	MSG_SOURCE = "_source"
	MSG_TARGET = "_target"
	MSG_HANDLE = "_handle"

	MSG_SESSID   = "sessid"
	MSG_USERNAME = "user.name"
	MSG_USERROLE = "user.role"

	MSG_RIVER = "sess.river"
	MSG_STORM = "sess.storm"
)
const (
	AAA_ROLE = "role"
	AAA_USER = "user"
	AAA_SESS = "sess"
)
const (
	WEB_PORT = ":9020"
	WEB_SESS = "sessid"
	WEB_TMPL = "render"

	WEB_LOGIN = "_login"
	WEB_SPIDE = "spide"
	WEB_SERVE = "serve"
	WEB_SPACE = "space"
	WEB_DREAM = "dream"
	WEB_FAVOR = "favor"
	WEB_CACHE = "cache"
	WEB_STORY = "story"
	WEB_SHARE = "share"
	WEB_ROUTE = "route"
	WEB_PROXY = "proxy"
)
const (
	LOG_CMD   = "cmd"
	LOG_INFO  = "info"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"
	LOG_TRACE = "trace"

	LOG_BEGIN = "begin"
	LOG_START = "start"
	LOG_BENCH = "bench"
	LOG_CLOSE = "close"
)
const (
	GDB_SIGNAL = "signal"
	GDB_TIMER  = "timer"
	GDB_EVENT  = "event"
)

const (
	CHAT_GROUP = "group"
)

var Alias = map[string]string{
	CTX_CONFIG:  "ctx.config",
	CTX_COMMAND: "ctx.command",
	CTX_CONTEXT: "ctx.context",

	CLI_RUNTIME: "cli.runtime",
	CLI_SYSTEM:  "cli.system",

	AAA_ROLE: "aaa.role",
	AAA_USER: "aaa.user",
	AAA_SESS: "aaa.sess",

	WEB_SPIDE: "web.spide",
	WEB_SERVE: "web.serve",
	WEB_SPACE: "web.space",
	WEB_DREAM: "web.dream",
	WEB_FAVOR: "web.favor",
	WEB_CACHE: "web.cache",
	WEB_STORY: "web.story",
	WEB_SHARE: "web.share",
	WEB_ROUTE: "web.route",
	WEB_PROXY: "web.proxy",

	GDB_SIGNAL: "gdb.signal",
	GDB_TIMER:  "gdb.timer",
	GDB_EVENT:  "gdb.event",

	CHAT_GROUP: "web.chat.group",

	"note": "web.wiki.note",
}
