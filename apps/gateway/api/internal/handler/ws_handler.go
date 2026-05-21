package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"im/apps/gateway/api/internal/hub"
	"im/apps/gateway/api/internal/logic"
	"im/apps/gateway/api/internal/middleware"
	"im/apps/gateway/api/internal/protocol"
	"im/apps/gateway/api/internal/svc"
	"im/pkg/code"
)

func WSHandler(svcCtx *svc.ServiceContext, h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := middleware.AuthenticateWS(r, svcCtx.Config.Auth.AccessSecret)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := h.Upgrader().Upgrade(w, r, nil)
		if err != nil {
			return
		}
		wsConf := h.WsConfig()
		conn.SetReadLimit(wsConf.MaxMessageSize())

		client := hub.NewConnection(conn)
		session := hub.NewSession()
		ctx := r.Context()

		h.BindUser(uid, client)
		session.SetAuth(uid, client)
		h.SetOnline(ctx, uid)
		_ = h.Send(client, protocol.NewAuthOK(uid))

		connCtx, connCancel := context.WithCancel(ctx)
		defer connCancel()
		connHB := hub.NewConnectionHeartbeat(h.WsConfig().HeartbeatMaxMissCount())
		h.RunConnectionHeartbeat(connCtx, uid, conn, client, connHB)

		defer func() {
			h.Unregister(uid, client)
		}()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			h.RefreshReadDeadline(conn)
			if int64(len(data)) > wsConf.MaxMessageSize() {
				_ = h.Send(client, protocol.NewErrorOut(code.GatewayMessageTooBig))
				continue
			}

			var frame protocol.InFrame
			if err := json.Unmarshal(data, &frame); err != nil {
				_ = h.Send(client, protocol.NewErrorOut(code.GatewayInvalidFrame, "invalid json"))
				continue
			}
			if err := frame.Validate(); err != nil {
				_ = h.Send(client, protocol.NewErrorOut(code.GatewayInvalidFrame, err.Error()))
				continue
			}
			// 任意上行帧续期 online:{uid}，避免仅 ping 导致 TTL 过期仍走 offline-push
			h.TouchOnline(ctx, uid)

			switch frame.Type {
			case protocol.TypeSend:
				out, errOut := logic.NewWSSendLogic(ctx, svcCtx).Send(frame, session)
				if errOut != nil {
					_ = h.Send(client, *errOut)
					continue
				}
				_ = h.Send(client, out)

			case protocol.TypePing:
				connHB.Ack()
				out := logic.NewWSPingLogic(ctx, h).Ping(session)
				_ = h.Send(client, out)
			}
		}
	}
}
