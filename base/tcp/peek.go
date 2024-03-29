package tcp

import (
	"bytes"
	"net"
	"net/http"
	"strings"

	kit "shylinux.com/x/toolkits"
)

type Buf struct {
	buf []byte
}
type PeekConn struct {
	net.Conn
	*Buf
}

func (s PeekConn) Read(b []byte) (n int, err error) {
	if len(s.buf) == 0 {
		return s.Conn.Read(b)
	}
	if len(s.buf) < len(b) {
		copy(b, s.buf)
		s.buf = s.buf[:0]
		return len(s.buf), nil
	}
	copy(b, s.buf)
	s.buf = s.buf[len(b):]
	return len(b), nil
}
func (s PeekConn) Peek(n int) (res []byte) {
	b := make([]byte, n)
	_n, _ := s.Conn.Read(b)
	s.Buf.buf = b[:_n]
	return b[:_n]
}
func (s PeekConn) IsHTTP() bool {
	if head := s.Peek(4); bytes.Equal(head, []byte("GET ")) {
		return true
	}
	return false
}
func (s PeekConn) Redirect(status int, location string) {
	DF, NL := ": ", "\r\n"
	s.Conn.Write([]byte(strings.Join([]string{
		kit.Format("HTTP/1.1 %d %s", status, http.StatusText(status)),
		kit.JoinKV(DF, NL, "Location", location, "Content-Length", "0"),
	}, NL) + NL + NL))
}
func NewPeekConn(c net.Conn) PeekConn { return PeekConn{Conn: c, Buf: &Buf{}} }
