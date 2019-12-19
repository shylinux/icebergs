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
)
const (
	GDB_SIGNAL = "signal"
	GDB_TIMER  = "timer"
	GDB_EVENT  = "event"
)
const (
	WEB_PORT = ":9020"
	WEB_SESS = "sessid"
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
	GDB_SIGNAL: "gdb.signal",
	GDB_TIMER:  "gdb.timer",
	GDB_EVENT:  "gdb.event",
}
