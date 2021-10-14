package ice

const (
	SP = " "
	NL = "\n"
	PT = "."

	OK      = "ok"
	TRUE    = "true"
	FALSE   = "false"
	SUCCESS = "success"
	FAILURE = "failure"

	INIT = "init"
	LOAD = "load"
	AUTO = "auto"
	HELP = "help"
	SAVE = "save"
	EXIT = "exit"

	SHY = "shy"
	DEV = "dev"
	OPS = "ops"
	ICE = "ice"
	RUN = "run"

	POD = "pod"
	CTX = "ctx"
	CMD = "cmd"
	ARG = "arg"
	OPT = "opt"

	SRC = "src"
	ETC = "etc"
	BIN = "bin"
	VAR = "var"
	USR = "usr"
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
	ICEBERGS = "icebergs"
	INTSHELL = "intshell"
	CONTEXTS = "contexts"

	REQUIRE = "require"
	INSTALL = "install"
	PUBLISH = "publish"
)
const ( // DIR
	USR_VOLCANOS = "usr/volcanos"
	USR_LEARNING = "usr/learning"
	USR_ICEBERGS = "usr/icebergs"
	USR_INTSHELL = "usr/intshell"
	USR_INSTALL  = "usr/install"
	USR_PUBLISH  = "usr/publish"

	USR_LOCAL        = "usr/local"
	USR_LOCAL_WORK   = "usr/local/work"
	USR_LOCAL_IMAGE  = "usr/local/image"
	USR_LOCAL_RIVER  = "usr/local/river"
	USR_LOCAL_DAEMON = "usr/local/daemon"
	USR_LOCAL_EXPORT = "usr/local/export"

	FAVICON  = "favicon.ico"
	PROTO_JS = "proto.js"
	FRAME_JS = "frame.js"
	INDEX_JS = "index.js"
	ORDER_JS = "order.js"
	ORDER_SH = "order.sh"
	INDEX_SH = "index.sh"

	VAR_RUN      = "var/run"
	VAR_TMP      = "var/tmp"
	VAR_LOG      = "var/log"
	VAR_CONF     = "var/conf"
	VAR_DATA     = "var/data"
	VAR_FILE     = "var/file"
	VAR_PROXY    = "var/proxy"
	VAR_TRASH    = "var/trash"
	BIN_ICE_SH   = "bin/ice.sh"
	BIN_ICE_BIN  = "bin/ice.bin"
	BIN_BOOT_LOG = "bin/boot.log"
	ETC_INIT_SHY = "etc/init.shy"
	ETC_EXIT_SHY = "etc/exit.shy"
	ETC_MISS_SH  = "etc/miss.sh"
	ETC_PATH     = "etc/path"

	SRC_HELP       = "src/help"
	SRC_MAIN_SHY   = "src/main.shy"
	SRC_MAIN_GO    = "src/main.go"
	SRC_VERSION_GO = "src/version.go"
	SRC_BINPACK_GO = "src/binpack.go"
	MAKEFILE       = "makefile"
	GO_MOD         = "go.mod"
	GO_SUM         = "go.sum"
)
const ( // MSG
	MSG_DETAIL = "detail"
	MSG_OPTION = "option"
	MSG_APPEND = "append"
	MSG_RESULT = "result"

	MSG_ALIAS  = "_alias"
	MSG_SCRIPT = "_script"
	MSG_SOURCE = "_source"
	MSG_TARGET = "_target"
	MSG_HANDLE = "_handle"
	MSG_RENDER = "_render"
	MSG_OUTPUT = "_output"
	MSG_ARGS   = "_args"

	MSG_UPLOAD = "_upload"
	MSG_DAEMON = "_daemon"
	MSG_ACTION = "_action"
	MSG_STATUS = "_status"

	MSG_DISPLAY = "_display"
	MSG_PROCESS = "_process"

	MSG_CMDS   = "cmds"
	MSG_FIELDS = "fields"
	MSG_SESSID = "sessid"
	MSG_DOMAIN = "domain"
	MSG_OPTS   = "_option"

	MSG_USERIP   = "user.ip"
	MSG_USERUA   = "user.ua"
	MSG_USERWEB  = "user.web"
	MSG_USERPOD  = "user.pod"
	MSG_USERADDR = "user.addr"
	MSG_USERDATA = "user.data"
	MSG_USERNICK = "user.nick"
	MSG_USERNAME = "user.name"
	MSG_USERROLE = "user.role"
	MSG_USERZONE = "user.zone"
	MSG_LANGUAGE = "user.lang"

	MSG_TITLE = "sess.title"
	MSG_TOPIC = "sess.topic"
	MSG_RIVER = "sess.river"
	MSG_STORM = "sess.storm"
	MSG_LOCAL = "sess.local"
	MSG_TOAST = "sess.toast"

	CACHE_LIMIT  = "cache.limit"
	CACHE_BEGIN  = "cache.begin"
	CACHE_COUNT  = "cache.count"
	CACHE_OFFEND = "cache.offend"
)
const ( // RENDER
	RENDER_RAW      = "_raw"
	RENDER_VOID     = "_void"
	RENDER_RESULT   = "_result"
	RENDER_ANCHOR   = "_anchor"
	RENDER_BUTTON   = "_button"
	RENDER_SCRIPT   = "_script"
	RENDER_QRCODE   = "_qrcode"
	RENDER_IMAGES   = "_images"
	RENDER_VIDEOS   = "_videos"
	RENDER_TEMPLATE = "_template"
	RENDER_REDIRECT = "_redirect"
	RENDER_DOWNLOAD = "_download"
)
const ( // PROCESS
	PROCESS_LOCATION = "_location"
	PROCESS_REFRESH  = "_refresh"
	PROCESS_REWRITE  = "_rewrite"
	PROCESS_DISPLAY  = "_display"
	PROCESS_FIELD    = "_field"
	PROCESS_INNER    = "_inner"
	PROCESS_AGAIN    = "_again"

	PROCESS_OPEN = "_open"
	PROCESS_HOLD = "_hold"
	PROCESS_BACK = "_back"
	PROCESS_GROW = "_grow"

	FIELD_PREFIX = "_prefix"
)
const ( // Err
	ErrWarn     = "warn: "
	ErrExists   = "exists: "
	ErrExpire   = "expire: "
	ErrNotLogin = "not login: "
	ErrNotFound = "not found: "
	ErrNotRight = "not right: "
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
