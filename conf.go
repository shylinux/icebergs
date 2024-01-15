package ice

import kit "shylinux.com/x/toolkits"

const (
	TB = "\t"
	SP = " "
	DF = ":"
	EQ = "="
	AT = "@"
	QS = "?"
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
	WINDOWS = "windows"

	HTTPS = "https"
	HTTP  = "http"
	MAIL  = "mail"
	DEMO  = "demo"
	HELP  = "help"
	MAIN  = "main"
	AUTO  = "auto"
	LIST  = "list"
	BACK  = "back"

	BASE = "base"
	CORE = "core"
	MISC = "misc"

	SHY = "shy"
	DEV = "dev"
	OPS = "ops"

	ICE = "ice"
	CAN = "can"
	APP = "app"

	POD = "pod"
	CMD = "cmd"
	ARG = "arg"
	ENV = "env"
	RUN = "run"
	RES = "res"
	ERR = "err"
)
const ( // REPOS
	CONTEXTS = "contexts"
	INTSHELL = "intshell"
	VOLCANOS = "volcanos"
	LEARNING = "learning"
	TOOLKITS = "toolkits"
	ICEBERGS = "icebergs"
	RELEASE  = "release"
	MATRIX   = "matrix"
	ICONS    = "icons"

	INSTALL = "install"
	REQUIRE = "require"
	PUBLISH = "publish"
	PORTAL  = "portal"
	LOCAL   = "local"
)
const ( // MOD
	MOD_DIR       = 0750
	MOD_FILE      = 0640
	MOD_BUFS      = 4096
	MOD_DATE      = kit.MOD_DATE
	MOD_TIME      = kit.MOD_TIME
	MOD_TIMES     = kit.MOD_TIMES
	MOD_TIME_ONLY = "15:04:05"
)

const ( // DIR
	SRC = "src"
	ETC = "etc"
	BIN = "bin"
	VAR = "var"
	USR = "usr"

	JSON = "json"
	CSV  = "csv"
	SH   = "sh"
	GO   = "go"
	JS   = "js"
	SVG  = "svg"
	CSS  = "css"
	HTML = "html"

	LIB    = "lib"
	PAGE   = "page"
	PANEL  = "panel"
	PLUGIN = "plugin"
	STORY  = "story"

	INDEX_CSS = "index.css"
	CONST_JS  = "const.js"
	PROTO_JS  = "proto.js"
	FRAME_JS  = "frame.js"
	INDEX_SH  = "index.sh"

	PLUGIN_INPUT    = "/plugin/input/"
	PLUGIN_LOCAL    = "/plugin/local/"
	PLUGIN_STORY    = "/plugin/story/"
	PLUGIN_TABLE_JS = "/plugin/table.js"

	ISH_PLUGED = ".ish/pluged/"

	USR_LOCAL    = "usr/local/"
	USR_PORTAL   = "usr/portal/"
	USR_INSTALL  = "usr/install/"
	USR_REQUIRE  = "usr/require/"
	USR_PUBLISH  = "usr/publish/"
	USR_RELEASE  = "usr/release/"
	USR_ICEBERGS = "usr/icebergs/"
	USR_TOOLKITS = "usr/toolkits/"
	USR_VOLCANOS = "usr/volcanos/"
	USR_LEARNING = "usr/learning/"
	USR_INTSHELL = "usr/intshell/"
	USR_PROGRAM  = "usr/program/"
	USR_GEOAREA  = "usr/geoarea/"
	USR_ICONS    = "usr/icons/"

	USR_LOCAL_GO       = "usr/local/go/"
	USR_LOCAL_GO_BIN   = "usr/local/go/bin/"
	USR_LOCAL_BIN      = "usr/local/bin/"
	USR_LOCAL_LIB      = "usr/local/lib/"
	USR_LOCAL_WORK     = "usr/local/work/"
	USR_LOCAL_REPOS    = "usr/local/repos/"
	USR_LOCAL_IMAGE    = "usr/local/image/"
	USR_LOCAL_EXPORT   = "usr/local/export/"
	USR_LOCAL_DAEMON   = "usr/local/daemon/"
	VAR_DATA_IMPORTANT = "var/data/.important"
	VAR_LOG_BOOT_LOG   = "var/log/boot.log"
	VAR_LOG_ICE_PID    = "var/log/ice.pid"

	VAR_LOG            = "var/log/"
	VAR_TMP            = "var/tmp/"
	VAR_CONF           = "var/conf/"
	VAR_DATA           = "var/data/"
	VAR_FILE           = "var/file/"
	VAR_PROXY          = "var/proxy/"
	VAR_TRASH          = "var/trash/"
	BIN_ICE_BIN        = "bin/ice.bin"
	ETC_INIT_SHY       = "etc/init.shy"
	ETC_LOCAL_SHY      = "etc/local.shy"
	ETC_EXIT_SHY       = "etc/exit.shy"
	ETC_MISS_SH        = "etc/miss.sh"
	ETC_PATH           = "etc/path"
	SRC_MAIN_SH        = "src/main.sh"
	SRC_MAIN_SHY       = "src/main.shy"
	SRC_MAIN_HTML      = "src/main.html"
	SRC_MAIN_ICO       = "src/main.ico"
	SRC_MAIN_CSS       = "src/main.css"
	SRC_MAIN_JS        = "src/main.js"
	SRC_MAIN_GO        = "src/main.go"
	SRC_WEBVIEW_GO     = "src/webview.go"
	SRC_VERSION_GO     = "src/version.go"
	SRC_BINPACK_GO     = "src/binpack.go"
	SRC_BINPACK_USR_GO = "src/binpack_usr.go"
	SRC_TEMPLATE       = "src/template/"
	SRC_SCRIPT         = "src/script/"
	USR_SCRIPT         = "usr/script/"
	README_MD          = "README.md"
	MAKEFILE           = "Makefile"
	LICENSE            = "LICENSE"
	GO_MOD             = "go.mod"
	GO_SUM             = "go.sum"
	ICE_BIN            = "ice.bin"
	CAN_PLUGIN         = "can._plugin"
)
const ( // MSG
	MSG_CMDS   = "cmds"
	MSG_FIELDS = "fields"
	MSG_METHOD = "method"
	MSG_SESSID = "sessid"
	MSG_DEBUG  = "debug"

	MSG_DETAIL = "detail"
	MSG_OPTION = "option"
	MSG_APPEND = "append"
	MSG_RESULT = "result"

	MSG_OPTS   = "_option"
	MSG_UPLOAD = "_upload"
	MSG_SOURCE = "_source"
	MSG_TARGET = "_target"
	MSG_ACTION = "_action"
	MSG_STATUS = "_status"

	MSG_SPACE  = "_space"
	MSG_INDEX  = "_index"
	MSG_SCRIPT = "_script"
	MSG_OUTPUT = "_output"
	MSG_ARGS   = "_args"

	MSG_PROCESS = "_process"
	MSG_DISPLAY = "_display"
	MSG_TOOLKIT = "_toolkit"

	MSG_USERIP   = "user.ip"
	MSG_USERUA   = "user.ua"
	MSG_USERWEB  = "user.web"
	MSG_USERPOD  = "user.pod"
	MSG_USERHOST = "user.host"
	MSG_USERADDR = "user.addr"
	MSG_USERDATA = "user.data"
	MSG_USERNICK = "user.nick"
	MSG_USERNAME = "user.name"
	MSG_USERROLE = "user.role"
	MSG_USERZONE = "user.zone"
	MSG_LANGUAGE = "user.lang"

	MSG_BG      = "sess.bg"
	MSG_FG      = "sess.fg"
	MSG_COST    = "sess.cost"
	MSG_MODE    = "sess.mode"
	MSG_THEME   = "sess.theme"
	MSG_TITLE   = "sess.title"
	MSG_RIVER   = "sess.river"
	MSG_STORM   = "sess.storm"
	MSG_COUNT   = "sess.count"
	MSG_DAEMON  = "sess.daemon"
	MSG_REFERER = "sess.referer"
	MSG_FILES   = "file.system"
	MSG_CHECKER = "aaa.checker"
	YAC_MESSAGE = "yac.message"
	YAC_STACK   = "yac.stack"
	SSH_ALIAS   = "ssh.alias"
	SSH_TARGET  = "ssh.target"
	LOG_DISABLE = "log.disable"
	LOG_TRACEID = "log.id"

	TOAST_DURATION = "toast.duration"
	TABLE_CHECKBOX = "table.checkbox"
	TCP_DOMAIN     = "tcp_domain"
)
const ( // RENDER
	RENDER_BUTTON = "_button"
	RENDER_ANCHOR = "_anchor"
	RENDER_QRCODE = "_qrcode"
	RENDER_IMAGES = "_images"
	RENDER_VIDEOS = "_videos"
	RENDER_AUDIOS = "_audios"
	RENDER_IFRAME = "_iframe"
	RENDER_SCRIPT = "_script"

	RENDER_STATUS   = "_status"
	RENDER_REDIRECT = "_redirect"
	RENDER_DOWNLOAD = "_download"
	RENDER_TEMPLATE = "_template"
	RENDER_RESULT   = "_result"
	RENDER_JSON     = "_json"
	RENDER_VOID     = "_void"
	RENDER_RAW      = "_raw"
)
const ( // PROCESS
	PROCESS_COOKIE   = "_cookie"
	PROCESS_SESSION  = "_session"
	PROCESS_LOCATION = "_location"
	PROCESS_REPLACE  = "_replace"
	PROCESS_HISTORY  = "_history"
	PROCESS_CONFIRM  = "_confirm"
	PROCESS_REFRESH  = "_refresh"
	PROCESS_REWRITE  = "_rewrite"
	PROCESS_DISPLAY  = "_display"

	PROCESS_FIELD = "_field"
	PROCESS_FLOAT = "_float"
	PROCESS_INNER = "_inner"
	PROCESS_AGAIN = "_again"
	PROCESS_HOLD  = "_hold"
	PROCESS_BACK  = "_back"
	PROCESS_RICH  = "_rich"
	PROCESS_GROW  = "_grow"
	PROCESS_OPEN  = "_open"
	PROCESS_CLOSE = "_close"

	PROCESS_ARG   = "_arg"
	FIELD_PREFIX  = "_prefix"
	FIELDS_DETAIL = "detail"
)
const ( // CTX
	CTX_FOLLOW = "follow"

	CTX_BEGIN = "begin"
	CTX_START = "start"
	CTX_SERVE = "serve"
	CTX_CLOSE = "close"

	CTX_INIT  = "_init"
	CTX_OPEN  = "_open"
	CTX_EXIT  = "_exit"
	CTX_ICONS = "_icons"
	CTX_TRANS = "_trans"
	CTX_TITLE = "_title"
)
const ( // LOG
	LOG_CMDS  = "cmds"
	LOG_AUTH  = "auth"
	LOG_COST  = "cost"
	LOG_INFO  = "info"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"
	LOG_DEBUG = "debug"
)
const ( // Err
	ErrWarn = "warn: "

	ErrNotLogin = "not login: "
	ErrNotRight = "not right: "
	ErrNotAllow = "not allow: "
	ErrNotValid = "not valid: "
	ErrNotFound = "not found: "
	ErrNotStart = "not start: "

	ErrAlreadyExists = "already exists: "
	ErrNotImplement  = "not implement: "
	ErrTooDeepCount  = "too deep count: "
)
const ( // ctx
	COMMAND = "command"
	ACTION  = "action"
	STYLE   = "style"
	INDEX   = "index"
)
const ( // mdb
	SEARCH = "search"
	INPUTS = "inputs"
	CREATE = "create"
	SELECT = "select"

	KEY   = "key"
	FIELD = "field"
	VALUE = "value"
	EXTRA = "extra"
	META  = "meta"
	TIME  = "time"
	HASH  = "hash"
	TYPE  = "type"
	NAME  = "name"
	TEXT  = "text"
	LINK  = "link"
)
const ( // web
	SERVE = "serve"
	SPACE = "space"

	THEME = "theme"
	TITLE = "title"
)
const ( // gdb
	EVENT   = "event"
	ROUTINE = "routine"
)
const ( // nfs
	SIZE   = "size"
	SOURCE = "source"
	SCRIPT = "script"
)
const ( // cli
	FOREVER = "forever"
	SYSTEM  = "system"
	START   = "start"
)
const ( // log
	DEBUG = "debug"
)
const ( // ice
	CTX = "ctx"
	MDB = "mdb"
	WEB = "web"
	AAA = "aaa"
	LEX = "lex"
	YAC = "yac"
	SSH = "ssh"
	GDB = "gdb"
	TCP = "tcp"
	NFS = "nfs"
	CLI = "cli"
	LOG = "log"
)
const ( // env
	LOG_TRACE = "log_trace"
)
