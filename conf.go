package ice

const ( // ICE
	ICE_CHAN = 10
	ICE_INIT = "_init"
	ICE_EXIT = "_exit"
	ICE_DATE = "2006-01-02"
	ICE_TIME = "2006-01-02 15:04:05"

	ICE_BEGIN = "begin"
	ICE_START = "start"
	ICE_SERVE = "serve"
	ICE_CLOSE = "close"
)
const ( // MSG
	MSG_DETAIL = "detail"
	MSG_OPTION = "option"
	MSG_APPEND = "append"
	MSG_RESULT = "result"

	MSG_SOURCE = "_source"
	MSG_TARGET = "_target"
	MSG_HANDLE = "_handle"
	MSG_ACTION = "_action"
	MSG_OUTPUT = "_output"
	MSG_ARGS   = "_args"

	MSG_STDOUT = "_stdout"
	MSG_PROMPT = "_prompt"
	MSG_ALIAS  = "_alias"

	MSG_SESSID   = "sessid"
	MSG_USERIP   = "user.ip"
	MSG_USERUA   = "user.ua"
	MSG_USERURL  = "user.url"
	MSG_USERPOD  = "user.pod"
	MSG_USERWEB  = "user.web"
	MSG_USERNICK = "user.nick"
	MSG_USERNAME = "user.name"
	MSG_USERROLE = "user.role"
	MSG_USERADDR = "user.addr"

	MSG_RIVER = "sess.river"
	MSG_STORM = "sess.storm"
)

const ( // CTX
	CTX_STATUS  = "status"
	CTX_STREAM  = "stream"
	CTX_FOLLOW  = "follow"
	CTX_CONFIG  = "config"
	CTX_COMMAND = "command"
	CTX_CONTEXT = "context"
)
const ( // CLI
	CLI_RUNTIME = "runtime"
	CLI_SYSTEM  = "system"
	CLI_DAEMON  = "daemon"
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

	WEB_MASTER = "master"
	WEB_MYSELF = "myself"
	WEB_BETTER = "better"
	WEB_SERVER = "server"
	WEB_WORKER = "worker"
)
const ( // LOG
	// 数据
	LOG_ENABLE = "enable"
	LOG_IMPORT = "import"
	LOG_EXPORT = "export"
	LOG_CREATE = "create"
	LOG_REMOVE = "remove"
	LOG_INSERT = "insert"
	LOG_DELETE = "delete"
	LOG_SELECT = "select"
	LOG_MODIFY = "modify"

	// 事件
	LOG_LISTEN = "listen"
	LOG_SIGNAL = "signal"
	LOG_EVENTS = "events"
	LOG_TIMERS = "timers"

	// 状态
	LOG_BEGIN = "begin"
	LOG_START = "start"
	LOG_SERVE = "serve"
	LOG_CLOSE = "close"

	// 权限
	LOG_LOGIN  = "login"
	LOG_LOGOUT = "logout"

	// 分类
	LOG_CMDS  = "cmds"
	LOG_COST  = "cost"
	LOG_INFO  = "info"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"
	LOG_TRACE = "trace"
)
const ( // SSH
	SSH_SOURCE = "source"
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
	CHAT_CREATE = "chat.create"
	MISS_CREATE = "miss.create"
	MIND_CREATE = "mind.create"
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
	APP_MIND    = "mind"
	APP_MISS    = "miss"
	APP_SEARCH  = "search"
	APP_COMMEND = "commend"
)
const ( // ROLE
	ROLE_ROOT = "root"
	ROLE_TECH = "tech"
	ROLE_VOID = "void"
)
const ( // TYPE
	TYPE_SPIDE = "spide"
	TYPE_SPACE = "space"
	TYPE_STORY = "story"

	TYPE_RIVER  = "river"
	TYPE_STORM  = "storm"
	TYPE_ACTION = "action"
	TYPE_ACTIVE = "active"

	TYPE_DRIVE = "drive"
	TYPE_SHELL = "shell"
	TYPE_VIMRC = "vimrc"
	TYPE_TABLE = "table"
	TYPE_INNER = "inner"
	TYPE_MEDIA = "media"
)
const ( // CODE
	CODE_INSTALL = "_install"
	CODE_PREPARE = "_prepare"
	CODE_PROJECT = "_project"
)
const ( // CHAT
	CHAT_RIVER = "river"
	CHAT_STORM = "storm"
)
const ( // FAVOR
	FAVOR_CHAT  = "chat.init"
	FAVOR_TMUX  = "tmux.init"
	FAVOR_START = "favor.start"
)
const ( // STORY
	STORY_CATCH = "catch"
	STORY_INDEX = "index"
	STORY_TRASH = "trash"
	STORY_WATCH = "watch"

	STORY_STATUS  = "status"
	STORY_COMMIT  = "commit"
	STORY_BRANCH  = "branch"
	STORY_HISTORY = "history"

	STORY_PULL     = "pull"
	STORY_PUSH     = "push"
	STORY_UPLOAD   = "upload"
	STORY_DOWNLOAD = "download"
)
const ( // RENDER
	RENDER_VOID     = "_output"
	RENDER_OUTPUT   = "_output"
	RENDER_TEMPLATE = "_template"
	RENDER_DOWNLOAD = "_download"
	RENDER_RESULT   = "_result"
	RENDER_QRCODE   = "_qrcode"
)

var Alias = map[string]string{
	CTX_CONTEXT: "ctx.context",
	CTX_COMMAND: "ctx.command",
	CTX_CONFIG:  "ctx.config",

	CLI_RUNTIME: "cli.runtime",
	CLI_SYSTEM:  "cli.system",
	CLI_DAEMON:  "cli.daemon",
	SSH_SOURCE:  "ssh.source",

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

	APP_MISS:    "web.team.miss",
	APP_MIND:    "web.wiki.mind",
	APP_SEARCH:  "web.chat.search",
	APP_COMMEND: "web.chat.commend",

	"compile": "web.code.compile",
	"publish": "web.code.publish",
	"upgrade": "web.code.upgrade",
	"pprof":   "web.code.pprof",
}
