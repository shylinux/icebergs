package yac

import kit "shylinux.com/x/toolkits"

type Error struct {
	key      string
	detail   string
	fileline string
	Position
}

const (
	ERROR = "error"

	errNotImplement = "not implement: "
	errNotSupport   = "not support: "
	errNotValid     = "not valid: "
	errNotFound     = "not found: "
)

func errCommon(key string, arg ...Any) Error {
	detail := ""
	if len(arg) == 1 {
		switch v := arg[0].(type) {
		case string:
			detail = v
		default:
			kit.Format("%#v", v)
		}
	} else if len(arg) > 1 {
		detail = kit.Format(arg[0], arg[1:]...)
	}

	return Error{key: key, detail: detail, fileline: kit.FileLine(3, 100)}
}
func ErrNotImplement(arg ...Any) Error { return errCommon(errNotImplement, arg...) }
func ErrNotSupport(arg ...Any) Error   { return errCommon(errNotSupport, arg...) }
func ErrNotValid(arg ...Any) Error     { return errCommon(errNotValid, arg...) }
func ErrNotFound(arg ...Any) Error     { return errCommon(errNotFound, arg...) }
