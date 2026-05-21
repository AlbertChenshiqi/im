package hub

import "sync"

type Registry struct {
	mu      sync.RWMutex
	single  bool
	byUID   map[int64]map[uint64]*Client
}

func NewRegistry(singleMode bool) *Registry {
	return &Registry{
		single: singleMode,
		byUID:  make(map[int64]map[uint64]*Client),
	}
}

func (r *Registry) Register(uid int64, client *Client) (kicked []*Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	client.uid = uid
	conns := r.byUID[uid]
	if r.single && len(conns) > 0 {
		for _, old := range conns {
			kicked = append(kicked, old)
		}
		r.byUID[uid] = map[uint64]*Client{client.id: client}
		return kicked
	}

	if conns == nil {
		conns = make(map[uint64]*Client)
		r.byUID[uid] = conns
	}
	conns[client.id] = client
	return kicked
}

func (r *Registry) Unregister(uid int64, clientID uint64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	conns := r.byUID[uid]
	if conns == nil {
		return true
	}
	delete(conns, clientID)
	if len(conns) == 0 {
		delete(r.byUID, uid)
		return true
	}
	return false
}

func (r *Registry) Broadcast(uid int64, v any) int {
	r.mu.RLock()
	conns := r.byUID[uid]
	clients := make([]*Client, 0, len(conns))
	for _, c := range conns {
		clients = append(clients, c)
	}
	r.mu.RUnlock()

	n := 0
	for _, c := range clients {
		if c.Enqueue(v) {
			n++
		}
	}
	return n
}

func (r *Registry) HasConnections(uid int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byUID[uid]) > 0
}

func (r *Registry) ConnectionCount(uid int64) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byUID[uid])
}
