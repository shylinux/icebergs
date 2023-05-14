package websocket

import (
	"net"
	"net/http"
	"net/url"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/toolkits/task"
	"shylinux.com/x/websocket"
)

type Conn struct {
	*websocket.Conn
	lock task.Lock
}

func (c *Conn) WriteMessage(messageType int, data []byte) error {
	defer c.lock.Lock()()
	return c.Conn.WriteMessage(messageType, data)
}
func Upgrade(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	conn, e := websocket.Upgrade(w, r, nil, ice.MOD_BUFS, ice.MOD_BUFS)
	return &Conn{Conn: conn}, e
}
func NewClient(c net.Conn, u *url.URL) (*Conn, error) {
	conn, _, e := websocket.NewClient(c, u, nil, ice.MOD_BUFS, ice.MOD_BUFS)
	return &Conn{Conn: conn}, e
}
