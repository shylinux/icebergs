package ice

const ( // MOD
	MOD_SP   = " "
	MOD_NL   = "\n"
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

	INSTALL = "install"
	REQUIRE = "require"
	PUBLISH = "publish"

	SUCCESS = "success"
	FAILURE = "failure"
	FALSE   = "false"
	TRUE    = "true"
	OK      = "ok"
)
const ( // DIR
	USR_VOLCANOS = "usr/volcanos"
	USR_ICEBERGS = "usr/icebergs"
	USR_INTSHELL = "usr/intshell"
	USR_INSTALL  = "usr/install"
	USR_PUBLISH  = "usr/publish"
	USR_LOCAL    = "usr/local"

	PROTO_JS = "proto.js"
	FRAME_JS = "frame.js"
	INDEX_JS = "index.js"
	ORDER_JS = "order.js"
	ORDER_SH = "order.sh"
	INDEX_SH = "index.sh"

	VAR_TMP     = "var/tmp"
	VAR_RUN     = "var/run"
	VAR_LOG     = "var/log"
	VAR_CONF    = "var/conf"
	VAR_DATA    = "var/data"
	VAR_FILE    = "var/file"
	VAR_PROXY   = "var/proxy"
	VAR_TRASH   = "var/trash"
	BIN_ICE     = "bin/ice.sh"
	BIN_ICE_BIN = "bin/ice.bin"
	BIN_BOOTLOG = "bin/boot.log"
	ETC_INIT    = "etc/init.shy"
	ETC_EXIT    = "etc/exit.shy"
	ETC_MISS    = "etc/miss.sh"

	SRC_MAIN    = "src/main.shy"
	SRC_MAIN_GO = "src/main.go"
	SRC_VERSION = "src/version.go"
	SRC_BINPACK = "src/binpack.go"
	MAKEFILE    = "makefile"
	GO_MOD      = "go.mod"
	GO_SUM      = "go.sum"

	CTX_DEBUG = "ctx_debug"
	CTX_DEV   = "ctx_dev"
	CTX_PID   = "ctx_pid"
	CTX_LOG   = "ctx_log"
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
	MSG_OUTPUT = "_output"
	MSG_ARGS   = "_args"

	MSG_DAEMON = "_daemon"
	MSG_UPLOAD = "_upload"
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
	MSG_USERZONE = "user.zone"
	MSG_USERROLE = "user.role"

	MSG_TITLE = "sess.title"
	MSG_RIVER = "sess.river"
	MSG_STORM = "sess.storm"
	MSG_LOCAL = "sess.local"
	MSG_TOAST = "sess.toast"
)
const ( // RENDER
	RENDER_VOID     = "_void"
	RENDER_RESULT   = "_result"
	RENDER_ANCHOR   = "_anchor"
	RENDER_BUTTON   = "_button"
	RENDER_IMAGES   = "_images"
	RENDER_VIDEOS   = "_videos"
	RENDER_QRCODE   = "_qrcode"
	RENDER_SCRIPT   = "_script"
	RENDER_DOWNLOAD = "_download"
	RENDER_TEMPLATE = "_template"
)
const ( // PROCESS
	PROCESS_REFRESH = "_refresh"
	PROCESS_REWRITE = "_rewrite"
	PROCESS_FIELD   = "_field"
	PROCESS_INNER   = "_inner"

	PROCESS_HOLD = "_hold"
	PROCESS_BACK = "_back"
	PROCESS_GROW = "_grow"
	PROCESS_OPEN = "_open"

	FIELD_PREFIX = "_prefix"
)
const ( // LOG
	// 数据
	LOG_CREATE = "create"
	LOG_REMOVE = "remove"
	LOG_MODIFY = "modify"
	LOG_INSERT = "insert"
	LOG_DELETE = "delete"
	LOG_SELECT = "select"
	LOG_EXPORT = "export"
	LOG_IMPORT = "import"

	// 状态
	LOG_BEGIN = "begin"
	LOG_START = "start"
	LOG_SERVE = "serve"
	LOG_CLOSE = "close"

	// 分类
	LOG_CMDS  = "cmds"
	LOG_AUTH  = "auth"
	LOG_SEND  = "send"
	LOG_RECV  = "recv"
	LOG_COST  = "cost"
	LOG_INFO  = "info"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"
	LOG_DEBUG = "debug"
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
