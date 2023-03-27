package websocket

import (
	"net"
	"net/http"
	"net/url"

	ice "shylinux.com/x/icebergs"
	"shylinux.com/x/websocket"
)

type Conn struct{ *websocket.Conn }

func Upgrade(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	conn, e := websocket.Upgrade(w, r, nil, ice.MOD_BUFS, ice.MOD_BUFS)
	return &Conn{conn}, e
}
func NewClient(c net.Conn, u *url.URL) (*Conn, error) {
	conn, _, e := websocket.NewClient(c, u, nil, ice.MOD_BUFS, ice.MOD_BUFS)
	return &Conn{conn}, e
}
