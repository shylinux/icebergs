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

	HTTPS = "https"
	HTTP  = "http"
	AUTO  = "auto"
	LIST  = "list"
	BACK  = "back"

	BASE = "base"
	CORE = "core"
	MISC = "misc"

	SHY = "shy"
	COM = "com"
	DEV = "dev"
	OPS = "ops"
	ICE = "ice"
	CAN = "can"

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
	ICEBERGS = "icebergs"
	TOOLKITS = "toolkits"
	VOLCANOS = "volcanos"
	LEARNING = "learning"

	INSTALL = "install"
	REQUIRE = "require"
	PUBLISH = "publish"
	RELEASE = "release"
)
const ( // MOD
	MOD_DIR   = 0750
	MOD_FILE  = 0640
	MOD_BUFS  = 4096
	MOD_DATE  = kit.MOD_DATE
	MOD_TIME  = kit.MOD_TIME
	MOD_TIMES = kit.MOD_TIMES
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
	CSS  = "css"
	SVG  = "svg"
	HTML = "html"

	LIB    = "lib"
	PAGE   = "page"
	PANEL  = "panel"
	PLUGIN = "plugin"
	STORY  = "story"

	INDEX_CSS = "index.css"
	PROTO_JS  = "proto.js"
	FRAME_JS  = "frame.js"
	INDEX_SH  = "index.sh"

	FAVICON_ICO  = "/favicon.ico"
	PLUGIN_INPUT = "/plugin/input/"
	PLUGIN_LOCAL = "/plugin/local/"
	PLUGIN_STORY = "/plugin/story/"

	ISH_PLUGED   = ".ish/pluged/"
	USR_MODULES  = "usr/node_modules/"
	USR_INSTALL  = "usr/install/"
	USR_REQUIRE  = "usr/require/"
	USR_PUBLISH  = "usr/publish/"
	USR_RELEASE  = "usr/release/"
	USR_INTSHELL = "usr/intshell/"
	USR_ICEBERGS = "usr/icebergs/"
	USR_TOOLKITS = "usr/toolkits/"
	USR_VOLCANOS = "usr/volcanos/"
	USR_LEARNING = "usr/learning/"

	USR_LOCAL          = "usr/local/"
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

	VAR_LOG        = "var/log/"
	VAR_TMP        = "var/tmp/"
	VAR_CONF       = "var/conf/"
	VAR_DATA       = "var/data/"
	VAR_FILE       = "var/file/"
	VAR_PROXY      = "var/proxy/"
	VAR_TRASH      = "var/trash/"
	BIN_ICE_BIN    = "bin/ice.bin"
	ETC_INIT_SHY   = "etc/init.shy"
	ETC_LOCAL_SHY  = "etc/local.shy"
	ETC_EXIT_SHY   = "etc/exit.shy"
	ETC_MISS_SH    = "etc/miss.sh"
	ETC_PATH       = "etc/path"
	SRC_HELP       = "src/help/"
	SRC_DEBUG      = "src/debug/"
	SRC_RELEASE    = "src/release/"
	SRC_TEMPLATE   = "src/template/"
	SRC_MAIN_SHY   = "src/main.shy"
	SRC_MAIN_SH    = "src/main.sh"
	SRC_MAIN_JS    = "src/main.js"
	SRC_MAIN_GO    = "src/main.go"
	SRC_WEBVIEW_GO = "src/webview.go"
	SRC_VERSION_GO = "src/version.go"
	SRC_BINPACK_GO = "src/binpack.go"
	README_MD      = "README.md"
	MAKEFILE       = "Makefile"
	LICENSE        = "LICENSE"
	GO_MOD         = "go.mod"
	GO_SUM         = "go.sum"
	ICE_BIN        = "ice.bin"
	CAN_PLUGIN     = "can._plugin"
)
const ( // MSG
	MSG_CMDS   = "cmds"
	MSG_FIELDS = "fields"
	MSG_SESSID = "sessid"

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

	MSG_INDEX  = "_index"
	MSG_ALIAS  = "_alias"
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

	MSG_MODE    = "sess.mode"
	MSG_TITLE   = "sess.title"
	MSG_THEME   = "sess.theme"
	MSG_RIVER   = "sess.river"
	MSG_STORM   = "sess.storm"
	MSG_WIDTH   = "sess.width"
	MSG_HEIGHT  = "sess.height"
	MSG_DAEMON  = "sess.daemon"
	MSG_FILES   = "file.system"
	LOG_DISABLE = "log.disable"
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

	PROCESS_ARG   = "_arg"
	FIELD_PREFIX  = "_prefix"
	FIELDS_DETAIL = "detail"
)
const ( // CTX
	CTX_ARG    = "ctx_arg"
	CTX_DAEMON = "ctx_daemon"
	CTX_FOLLOW = "follow"

	CTX_BEGIN = "begin"
	CTX_START = "start"
	CTX_SERVE = "serve"
	CTX_CLOSE = "close"

	CTX_INIT = "_init"
	CTX_EXIT = "_exit"
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
	ErrNotValid = "not valid: "
	ErrNotFound = "not found: "
	ErrNotStart = "not start: "

	ErrNotImplement = "not implement: "
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
	HASH  = "hash"
	TIME  = "time"
	TYPE  = "type"
	NAME  = "name"
	TEXT  = "text"
	LINK  = "link"
)
const ( // web
	SERVE = "serve"
	SPACE = "space"

	TITLE = "title"
	THEME = "theme"
)
const ( // gdb
	EVENT   = "event"
	ROUTINE = "routine"
)
const ( // nfs
	SOURCE = "source"
	SCRIPT = "script"
)
const ( // cli
	SYSTEM = "system"
	START  = "start"
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
