package configzero

// Shared defaults for all go-zero services.
const (
	PostgresDSN = "postgres://im:im@localhost:5432/im?sslmode=disable"
	RedisAddr   = "localhost:6379"
	KafkaBroker = "localhost:9092"
	JWTSecret   = "dev-secret-change-in-production"
)

// RPC endpoint helpers (local dev).
const (
	UserRPC         = "127.0.0.1:20100"
	FriendRPC       = "127.0.0.1:20200"
	GroupRPC        = "127.0.0.1:20300"
	ConversationRPC = "127.0.0.1:20400"
	MessageRPC      = "127.0.0.1:20500"
	NotificationRPC = "127.0.0.1:20600"
	PushRPC         = "127.0.0.1:20700"
)
