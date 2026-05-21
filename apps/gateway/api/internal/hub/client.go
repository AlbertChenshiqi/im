package hub

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type outbound struct {
	ping bool
	data []byte
}

type Client struct {
	id   uint64
	uid  int64
	conn *websocket.Conn
	send chan outbound
	once sync.Once
}

func newClient(id uint64, conn *websocket.Conn) *Client {
	c := &Client{
		id:   id,
		conn: conn,
		send: make(chan outbound, 64),
	}
	go c.writePump()
	return c
}

func (c *Client) writePump() {
	const writeWait = 10 * time.Second
	for o := range c.send {
		_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		var err error
		if o.ping {
			err = c.conn.WriteMessage(websocket.PingMessage, nil)
		} else {
			err = c.conn.WriteMessage(websocket.TextMessage, o.data)
		}
		if err != nil {
			return
		}
	}
}

func (c *Client) Enqueue(v any) bool {
	b, err := json.Marshal(v)
	if err != nil {
		return false
	}
	select {
	case c.send <- outbound{data: b}:
		return true
	default:
		return false
	}
}

func (c *Client) EnqueuePing() bool {
	select {
	case c.send <- outbound{ping: true}:
		return true
	default:
		return false
	}
}

func (c *Client) Close() {
	c.once.Do(func() {
		close(c.send)
		_ = c.conn.Close()
	})
}

var clientIDSeq atomic.Uint64

func nextClientID() uint64 {
	return clientIDSeq.Add(1)
}
