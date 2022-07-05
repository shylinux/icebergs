package ice

const (
	TB = "\t"
	SP = " "
	DF = ":"
	EQ = "="
	AT = "@"
	PS = "/"
	PT = "."
	FS = ","
	NL = "\n"

	OK      = "ok"
	TRUE    = "true"
	FALSE   = "false"
	SUCCESS = "success"
	FAILURE = "failure"
	PROCESS = "process"
	RESTART = "restart"
	OF      = " of "

	INIT = "init"
	EXIT = "exit"
	SAVE = "save"
	LOAD = "load"

	AUTO = "auto"
	LIST = "list"
	BACK = "back"
	EXEC = "exec"

	SHOW = "show"
	PLAY = "play"
	HELP = "help"
	HTTP = "http"

	BASE = "base"
	CORE = "core"
	MISC = "misc"

	SHY = "shy"
	DEV = "dev"
	OPS = "ops"
	ICE = "ice"

	ENV = "env"
	RUN = "run"
	ERR = "err"

	POD = "pod"
	CTX = "ctx"
	CMD = "cmd"
	ARG = "arg"
	RES = "res"
)
const ( // MOD
	MOD_DIR  = 0750
	MOD_FILE = 0640

	MOD_CHAN = 16
	MOD_TICK = "1s"
	MOD_BUFS = 4096

	MOD_DATE = "2006-01-02"
	MOD_TIME = "2006-01-02 15:04:05"
)
const ( // REPOS
	VOLCANOS = "volcanos"
	LEARNING = "learning"
	ICEBERGS = "icebergs"
	TOOLKITS = "toolkits"
	INTSHELL = "intshell"
	CONTEXTS = "contexts"

	INSTALL = "install"
	RELEASE = "release"
	PUBLISH = "publish"
	REQUIRE = "require"
	DISPLAY = "display"
)
const ( // DIR
	SRC = "src"
	ETC = "etc"
	BIN = "bin"
	VAR = "var"
	USR = "usr"

	HTML = "html"
	CSS  = "css"
	JS   = "js"
	GO   = "go"
	SH   = "sh"
	SVG  = "svg"
	CSV  = "csv"
	JSON = "json"

	USR_VOLCANOS = "usr/volcanos"
	USR_LEARNING = "usr/learning"
	USR_ICEBERGS = "usr/icebergs"
	USR_TOOLKITS = "usr/toolkits"
	USR_INTSHELL = "usr/intshell"
	USR_INSTALL  = "usr/install"
	USR_RELEASE  = "usr/release"
	USR_PUBLISH  = "usr/publish"

	PLUGIN_INPUT = "/plugin/input"
	PLUGIN_STORY = "/plugin/story"
	PLUGIN_LOCAL = "/plugin/local"

	FAVICON  = "favicon.ico"
	PROTO_JS = "proto.js"
	FRAME_JS = "frame.js"
	INDEX_JS = "index.js"
	ORDER_JS = "order.js"
	ORDER_SH = "order.sh"
	INDEX_SH = "index.sh"

	USR_LOCAL        = "usr/local"
	USR_LOCAL_GO     = "usr/local/go"
	USR_LOCAL_GO_BIN = "usr/local/go/bin"
	USR_LOCAL_BIN    = "usr/local/bin"
	USR_LOCAL_LIB    = "usr/local/lib"
	USR_LOCAL_WORK   = "usr/local/work"
	USR_LOCAL_IMAGE  = "usr/local/image"
	USR_LOCAL_RIVER  = "usr/local/river"
	USR_LOCAL_DAEMON = "usr/local/daemon"
	USR_LOCAL_EXPORT = "usr/local/export"
	USR_LOCAL_REPOS  = "usr/local/repos"

	VAR_RUN       = "var/run"
	VAR_TMP       = "var/tmp"
	VAR_LOG       = "var/log"
	VAR_CONF      = "var/conf"
	VAR_DATA      = "var/data"
	VAR_FILE      = "var/file"
	VAR_PROXY     = "var/proxy"
	VAR_TRASH     = "var/trash"
	BIN_ICE_BIN   = "bin/ice.bin"
	BIN_BOOT_LOG  = "bin/boot.log"
	ETC_INIT_SHY  = "etc/init.shy"
	ETC_LOCAL_SHY = "etc/local.shy"
	ETC_EXIT_SHY  = "etc/exit.shy"
	ETC_MISS_SH   = "etc/miss.sh"
	ETC_PATH      = "etc/path"

	SRC_HELP       = "src/help"
	SRC_DEBUG      = "src/debug"
	SRC_RELEASE    = "src/release"
	SRC_MAIN_GO    = "src/main.go"
	SRC_MAIN_SHY   = "src/main.shy"
	SRC_MAIN_SVG   = "src/main.svg"
	SRC_VERSION_GO = "src/version.go"
	SRC_BINPACK_GO = "src/binpack.go"
	SRC_RELAY_GO   = "src/relay.go"
	README_MD      = "README.md"
	MAKEFILE       = "Makefile"
	LICENSE        = "LICENSE"
	ICE_BIN        = "ice.bin"
	GO_SUM         = "go.sum"
	GO_MOD         = "go.mod"
)
const ( // MSG
	MSG_DETAIL = "detail"
	MSG_OPTION = "option"
	MSG_APPEND = "append"
	MSG_RESULT = "result"

	MSG_CMDS   = "cmds"
	MSG_FIELDS = "fields"
	MSG_SESSID = "sessid"
	MSG_DOMAIN = "domain"
	MSG_OPTS   = "_option"

	MSG_ALIAS  = "_alias"
	MSG_SCRIPT = "_script"
	MSG_SOURCE = "_source"
	MSG_TARGET = "_target"
	MSG_HANDLE = "_handle"
	MSG_OUTPUT = "_output"
	MSG_ARGS   = "_args"

	MSG_UPLOAD = "_upload"
	MSG_DAEMON = "_daemon"
	MSG_ACTION = "_action"
	MSG_STATUS = "_status"

	MSG_DISPLAY = "_display"
	MSG_PROCESS = "_process"

	MSG_USERIP   = "user.ip"
	MSG_USERUA   = "user.ua"
	MSG_USERWEB  = "user.web"
	MSG_USERPOD  = "user.pod"
	MSG_USERADDR = "user.addr"
	MSG_USERDATA = "user.data"
	MSG_USERROLE = "user.role"
	MSG_USERNAME = "user.name"
	MSG_USERNICK = "user.nick"
	MSG_USERZONE = "user.zone"
	MSG_LANGUAGE = "user.lang"

	MSG_TITLE = "sess.title"
	MSG_TOPIC = "sess.topic"
	MSG_RIVER = "sess.river"
	MSG_STORM = "sess.storm"
	MSG_TOAST = "sess.toast"
	MSG_LOCAL = "sess.local"
)
const ( // RENDER
	RENDER_RAW      = "_raw"
	RENDER_VOID     = "_void"
	RENDER_JSON     = "_json"
	RENDER_STATUS   = "_status"
	RENDER_RESULT   = "_result"
	RENDER_ANCHOR   = "_anchor"
	RENDER_BUTTON   = "_button"
	RENDER_SCRIPT   = "_script"
	RENDER_QRCODE   = "_qrcode"
	RENDER_IMAGES   = "_images"
	RENDER_VIDEOS   = "_videos"
	RENDER_IFRAME   = "_iframe"
	RENDER_TEMPLATE = "_template"
	RENDER_REDIRECT = "_redirect"
	RENDER_DOWNLOAD = "_download"
)
const ( // PROCESS
	PROCESS_LOCATION = "_location"
	PROCESS_REPLACE  = "_replace"
	PROCESS_HISTORY  = "_history"
	PROCESS_CONFIRM  = "_confirm"
	PROCESS_REFRESH  = "_refresh"
	PROCESS_REWRITE  = "_rewrite"
	PROCESS_DISPLAY  = "_display"
	PROCESS_FIELD    = "_field"
	PROCESS_INNER    = "_inner"
	PROCESS_AGAIN    = "_again"

	PROCESS_HOLD = "_hold"
	PROCESS_BACK = "_back"
	PROCESS_RICH = "_rich"
	PROCESS_GROW = "_grow"
	PROCESS_OPEN = "_open"
	PROCESS_ARG  = "_arg"

	FIELD_PREFIX = "_prefix"
)
const ( // CACHE
	CACHE_LIMIT  = "cache.limit"
	CACHE_BEGIN  = "cache.begin"
	CACHE_COUNT  = "cache.count"
	CACHE_OFFEND = "cache.offend"
	CACHE_FILTER = "cache.filter"
	CACHE_VALUE  = "cache.value"
	CACHE_FIELD  = "cache.field"
	CACHE_DETAIL = "detail"
)
const ( // CTX
	CTX_FOLLOW = "follow"
	CTX_STATUS = "status"
	CTX_STREAM = "stream"

	CTX_BEGIN = "begin"
	CTX_START = "start"
	CTX_SERVE = "serve"
	CTX_CLOSE = "close"

	CTX_INIT = "_init"
	CTX_EXIT = "_exit"
)

const ( // LOG
	// 通用
	LOG_INFO  = "info"
	LOG_COST  = "cost"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"
	LOG_DEBUG = "debug"

	// 命令
	LOG_AUTH = "auth"
	LOG_CMDS = "cmds"
	LOG_SEND = "send"
	LOG_RECV = "recv"

	// 状态
	LOG_BEGIN = "begin"
	LOG_START = "start"
	LOG_SERVE = "serve"
	LOG_CLOSE = "close"

	// 数据
	LOG_CREATE = "create"
	LOG_REMOVE = "remove"
	LOG_INSERT = "insert"
	LOG_DELETE = "delete"
	LOG_MODIFY = "modify"
	LOG_SELECT = "select"
	LOG_EXPORT = "export"
	LOG_IMPORT = "import"
)
const ( // Err
	ErrWarn = "warn: "

	ErrNotLogin     = "not login: "
	ErrNotRight     = "not right: "
	ErrNotFound     = "not found: "
	ErrNotValid     = "not valid: "
	ErrNotStart     = "not start: "
	ErrNotImplement = "not implement: "
)

const ( // ice
	// CTX = "ctx"
	CLI = "cli"
	WEB = "web"
	AAA = "aaa"
	LEX = "lex"
	YAC = "yac"
	GDB = "gdb"
	LOG = "log"
	TCP = "tcp"
	NFS = "nfs"
	SSH = "ssh"
	MDB = "mdb"
)
const ( // ctx
	COMMAND = "command"
	ACTION  = "action"
	CONFIG  = "config"
	STYLE   = "style"
	INDEX   = "index"
	ARGS    = "args"
	INPUTS  = "inputs"
	FEATURE = "feature"
)
const ( // web
	SERVE = "serve"
	SPACE = "space"
	SPIDE = "spide"
	CACHE = "cache"

	WEBSITE = "website"
)
const ( // aaa
	ROLE     = "role"
	RIGHT    = "right"
	USERROLE = "userrole"
	ROOT     = "root"
)
const ( // mdb
	KEY    = "key"
	VALUE  = "value"
	SCRIPT = "script"
	LINK   = "link"

	META = "meta"
	HASH = "hash"
	TIME = "time"
	TYPE = "type"
	NAME = "name"
	TEXT = "text"
)
const ( // ssh
	SOURCE = "source"
)
const ( // gdb
	EVENT    = "event"
	LISTEN   = "listen"
	HAPPEN   = "happen"
	FILELINE = "fileline"
)
