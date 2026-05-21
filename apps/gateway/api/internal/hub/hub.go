package hub

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"

	"im/apps/gateway/api/internal/config"
	"im/apps/gateway/api/internal/svc"
	"im/pkg/redisclient"
)

type Hub struct {
	svc        *svc.ServiceContext
	registry   *Registry
	upgrader   websocket.Upgrader
	instanceID string
}

func New(s *svc.ServiceContext) *Hub {
	wsConf := s.Config.WebSocket
	allowed := wsConf.AllowedOrigins
	instanceID := os.Getenv("GATEWAY_INSTANCE_ID")
	if instanceID == "" {
		instanceID, _ = os.Hostname()
	}
	return &Hub{
		svc:        s,
		registry:   NewRegistry(wsConf.IsSingleConnection()),
		instanceID: instanceID,
		upgrader: websocket.Upgrader{
			CheckOrigin:     makeOriginChecker(allowed),
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
		},
	}
}

func makeOriginChecker(allowed []string) func(*http.Request) bool {
	if len(allowed) == 0 || (len(allowed) == 1 && allowed[0] == "*") {
		return func(*http.Request) bool { return true }
	}
	set := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		set[strings.TrimSpace(o)] = struct{}{}
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		_, ok := set[origin]
		return ok
	}
}

func (h *Hub) Upgrader() *websocket.Upgrader {
	return &h.upgrader
}

func (h *Hub) WsConfig() config.WebSocketConf {
	return h.svc.Config.WebSocket
}

func (h *Hub) Svc() *svc.ServiceContext {
	return h.svc
}

func NewConnection(conn *websocket.Conn) *Client {
	return newClient(nextClientID(), conn)
}

func (h *Hub) BindUser(uid int64, client *Client) {
	kicked := h.registry.Register(uid, client)
	for _, old := range kicked {
		old.Close()
	}
}

func (h *Hub) Unregister(uid int64, client *Client) {
	h.registry.Unregister(uid, client.id)
	client.Close()
	// 本机无连接时，从 online_gateways 移除本实例；仅当所有 gateway 均无连接才删 online:{uid}
	if !h.registry.HasConnections(uid) {
		_ = h.svc.Redis.RemoveGatewayPresence(context.Background(), uid, h.instanceID)
	}
}

func (h *Hub) Broadcast(uid int64, payload any) int {
	return h.registry.Broadcast(uid, payload)
}

func (h *Hub) Send(client *Client, payload any) bool {
	return client.Enqueue(payload)
}

func (h *Hub) TouchOnline(ctx context.Context, uid int64) {
	_ = h.svc.Redis.TouchGatewayPresence(ctx, uid, h.instanceID, h.WsConfig().OnlineTTLSeconds())
}

func (h *Hub) SetOnline(ctx context.Context, uid int64) {
	_ = h.svc.Redis.AddGatewayPresence(ctx, uid, h.instanceID, h.WsConfig().OnlineTTLSeconds())
}

func (h *Hub) InstanceID() string {
	return h.instanceID
}

// Session 单连接会话状态（读循环与 logic 共享）
type Session struct {
	mu     sync.RWMutex
	uid    int64
	client *Client
}

func NewSession() *Session {
	return &Session{}
}

func (s *Session) UID() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.uid
}

func (s *Session) SetAuth(uid int64, client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uid = uid
	s.client = client
}

func (s *Session) Client() *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.client
}

func (s *Session) IsAuthed() bool {
	return s.UID() > 0
}

// Redis helpers exposed for logic
func (h *Hub) Redis() *redisclient.Client {
	return h.svc.Redis
}
