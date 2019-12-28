package ice

const ( // ICE
	ICE_CHAN = 10
	ICE_INIT = "_init"
	ICE_EXIT = "_exit"
	ICE_DATE = "2006-01-02"
	ICE_TIME = "2006-01-02 15:04:05"
)
const ( // CTX
	CTX_STATUS  = "status"
	CTX_STREAM  = "stream"
	CTX_CONFIG  = "config"
	CTX_COMMAND = "command"
	CTX_CONTEXT = "context"
)
const ( // CLI
	CLI_RUNTIME = "runtime"
	CLI_SYSTEM  = "system"
)
const ( // MSG
	MSG_DETAIL = "detail"
	MSG_OPTION = "option"
	MSG_APPEND = "append"
	MSG_RESULT = "result"

	MSG_SOURCE = "_source"
	MSG_TARGET = "_target"
	MSG_HANDLE = "_handle"

	MSG_SESSID   = "sessid"
	MSG_USERIP   = "user.ip"
	MSG_USERUA   = "user.ua"
	MSG_USERURL  = "user.url"
	MSG_USERNAME = "user.name"
	MSG_USERROLE = "user.role"

	MSG_RIVER = "sess.river"
	MSG_STORM = "sess.storm"
)
const ( // AAA
	AAA_ROLE = "role"
	AAA_USER = "user"
	AAA_SESS = "sess"
)
const ( // WEB
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
	WEB_GROUP = "group"
	WEB_LABEL = "label"
)
const ( // LOG
	LOG_CMDS  = "cmds"
	LOG_COST  = "cost"
	LOG_INFO  = "info"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"
	LOG_TRACE = "trace"

	LOG_BEGIN = "begin"
	LOG_START = "start"
	LOG_BENCH = "bench"
	LOG_CLOSE = "close"
)
const ( // GDB
	GDB_SIGNAL = "signal"
	GDB_TIMER  = "timer"
	GDB_EVENT  = "event"

	SYSTEM_INIT = "system.init"

	SERVE_START = "serve.start"
	SERVE_CLOSE = "serve.close"
	SPACE_START = "space.start"
	SPACE_CLOSE = "space.close"
	DREAM_START = "dream.start"
	DREAM_CLOSE = "dream.close"

	USER_CREATE = "user.create"
)
const ( // MDB
	MDB_REDIS  = "redis"
	MDB_MYSQL  = "mysql"
	MDB_CREATE = "create"
	MDB_IMPORT = "import"
	MDB_EXPORT = "export"
	MDB_REMOVE = "remove"

	MDB_INESRT = "insert"
	MDB_UPDATE = "update"
	MDB_SELECT = "select"
	MDB_DELETE = "delete"
)

const ( // APP
	APP_NOTE = "note"
	APP_MISS = "miss"
)
const ( // ROLE
	ROLE_ROOT = "root"
	ROLE_TECH = "tech"
	ROLE_VOID = "void"
)
const ( // CHAT
	CHAT_RIVER = "river"
)
const ( // TYPE
	TYPE_SPACE = "space"
	TYPE_RIVER = "river"
	TYPE_STORM = "storm"

	TYPE_STORY = "story"
	TYPE_SHELL = "shell"
	TYPE_VIMRC = "vimrc"
	TYPE_TABLE = "table"
	TYPE_INNER = "inner"
	TYPE_MEDIA = "media"
)
const ( // FAVOR
	FAVOR_CHAT  = "chat.init"
	FAVOR_TMUX  = "tmux.init"
	FAVOR_RIVER = "river.init"
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
	WEB_GROUP: "web.group",
	WEB_LABEL: "web.label",

	GDB_SIGNAL: "gdb.signal",
	GDB_TIMER:  "gdb.timer",
	GDB_EVENT:  "gdb.event",

	MDB_REDIS:  "mdb.redis",
	MDB_MYSQL:  "mdb.mysql",
	MDB_CREATE: "mdb.create",
	MDB_IMPORT: "mdb.import",
	MDB_EXPORT: "mdb.export",
	MDB_REMOVE: "mdb.remove",

	MDB_INESRT: "mdb.insert",
	MDB_UPDATE: "mdb.update",
	MDB_SELECT: "mdb.select",
	MDB_DELETE: "mdb.delete",

	CHAT_RIVER: "web.chat.river",

	APP_NOTE: "web.wiki.note",
	APP_MISS: "web.team.miss",
}
