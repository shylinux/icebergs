package ice

const ( //MOD
	MOD_DIR  = 0750
	MOD_FILE = 0640

	MOD_CHAN = 16
	MOD_TICK = "1s"
	MOD_BUF  = 1024

	MOD_DATE = "2006-01-02"
	MOD_TIME = "2006-01-02 15:04:05"
)
const ( // MSG
	MSG_DETAIL = "detail"
	MSG_OPTION = "option"
	MSG_APPEND = "append"
	MSG_RESULT = "result"

	MSG_ALIAS  = "_alias"
	MSG_SOURCE = "_source"
	MSG_TARGET = "_target"
	MSG_HANDLE = "_handle"
	MSG_ACTION = "_action"
	MSG_OUTPUT = "_output"
	MSG_ARGS   = "_args"

	MSG_CMDS     = "cmds"
	MSG_SESSID   = "sessid"
	MSG_DOMAIN   = "domain"
	MSG_USERIP   = "user.ip"
	MSG_USERUA   = "user.ua"
	MSG_USERURL  = "user.url"
	MSG_USERWEB  = "user.web"
	MSG_USERPOD  = "user.pod"
	MSG_USERNICK = "user.nick"
	MSG_USERNAME = "user.name"
	MSG_USERZONE = "user.zone"
	MSG_USERROLE = "user.role"
	MSG_USERDATA = "user.data"
	MSG_USERADDR = "user.addr"

	MSG_RIVER  = "sess.river"
	MSG_STORM  = "sess.storm"
	MSG_ACTIVE = "sess.active"
	MSG_METHOD = "sess.method"
)
const ( // CTX
	CTX_STREAM = "stream"
	CTX_STATUS = "status"
	CTX_FOLLOW = "follow"

	CTX_INIT = "_init"
	CTX_EXIT = "_exit"

	CTX_BEGIN = "begin"
	CTX_START = "start"
	CTX_SERVE = "serve"
	CTX_CLOSE = "close"
)
const ( // LOG
	// 数据
	LOG_IMPORT = "import"
	LOG_EXPORT = "export"
	LOG_CREATE = "create"
	LOG_REMOVE = "remove"
	LOG_INSERT = "insert"
	LOG_DELETE = "delete"
	LOG_SELECT = "select"
	LOG_MODIFY = "modify"

	// 事件
	LOG_ENABLE = "enable"
	LOG_STATUS = "status"
	LOG_LISTEN = "listen"
	LOG_ACCEPT = "accept"
	LOG_FINISH = "finish"
	LOG_SIGNAL = "signal"
	LOG_TIMERS = "timers"
	LOG_EVENTS = "events"

	// 状态
	LOG_BEGIN = "begin"
	LOG_START = "start"
	LOG_SERVE = "serve"
	LOG_CLOSE = "close"

	// 分类
	LOG_CONF  = "conf"
	LOG_CMDS  = "cmds"
	LOG_AUTH  = "auth"
	LOG_COST  = "cost"
	LOG_INFO  = "info"
	LOG_WARN  = "warn"
	LOG_ERROR = "error"
	LOG_TRACE = "trace"
	LOG_DEBUG = "debug"
)
const ( // RENDER
	RENDER_VOID     = "_void"
	RENDER_OUTPUT   = "_output"
	RENDER_RESULT   = "_result"
	RENDER_QRCODE   = "_qrcode"
	RENDER_DOWNLOAD = "_download"
	RENDER_TEMPLATE = "_template"
)
