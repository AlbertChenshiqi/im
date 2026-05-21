// Load test for 10k-member group fan-out path (inbox-unread + Redis).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"im/pkg/convid"
	"im/pkg/db"
	"im/pkg/events"
	"im/pkg/kafka"
)

func main() {
	userAPI := flag.String("user-api", "http://localhost:10100", "user API base URL")
	groupID := flag.Int64("group", 0, "existing group id (0=create)")
	members := flag.Int("members", 1000, "member count for new group")
	msgs := flag.Int("msgs", 50, "messages to send")
	flag.Parse()

	ctx := context.Background()
	dsn := getenv("POSTGRES_DSN", "postgres://im:im@localhost:5432/im?sslmode=disable")
	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{Addr: getenv("REDIS_ADDR", "localhost:6379")})
	start := time.Now()

	token, uid := registerLogin(*userAPI, "load_"+fmt.Sprint(time.Now().Unix()), "loadtest")
	if *groupID == 0 {
		*groupID = createGroup("http://localhost:10300", token, *members)
		log.Printf("created group %d with %d members", *groupID, *members)
	}
	conv := convid.Group(*groupID)

	// simulate online subset
	onlineN := *members / 5
	if onlineN > 5000 {
		onlineN = 5000
	}
	for i := int64(1); i <= int64(onlineN); i++ {
		_ = rdb.Set(ctx, fmt.Sprintf("online:%d", i), "1", 0).Err()
	}

	writer := kafka.NewWriter([]string{getenv("KAFKA_BROKERS", "localhost:9092")}, events.TopicMessageSend)
	defer writer.Close()

	for i := 0; i < *msgs; i++ {
		evt := events.MessageSendEvent{
			MsgID:    int64(1_000_000 + i),
			ConvID:   conv,
			ConvType: "group",
			GroupID:  *groupID,
			SenderID: uid,
			Seq:      int64(i + 1),
			MsgType:  "text",
			Content:  fmt.Sprintf("load message %d", i),
			Ts:       time.Now().Unix(),
		}
		if err := kafka.PublishJSON(ctx, writer, conv, evt); err != nil {
			log.Fatalf("publish: %v", err)
		}
	}

	elapsed := time.Since(start)
	lag := measureKafkaLag(ctx, pool)
	log.Printf("done: %d msgs in %v, kafka consumer hint: check inbox-unread logs", *msgs, elapsed)
	log.Printf("metrics: redis online keys, unread keys — lag_estimate=%s", lag)
}

func registerLogin(base, user, pass string) (string, int64) {
	body, _ := json.Marshal(map[string]string{"username": user, "password": pass, "nickname": user})
	resp, err := http.Post(base+"/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var out struct {
		Token string `json:"token"`
		User  struct {
			ID int64 `json:"id"`
		} `json:"user"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out.Token, out.User.ID
}

func createGroup(base, token string, members int) int64 {
	ids := make([]int64, 0, members-1)
	for i := int64(2); i <= int64(members); i++ {
		ids = append(ids, i)
	}
	body, _ := json.Marshal(map[string]any{"name": "loadtest", "member_ids": ids})
	req, _ := http.NewRequest(http.MethodPost, base+"/v1/groups", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	var out struct {
		Group struct {
			ID int64 `json:"id"`
		} `json:"group"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out.Group.ID
}

func measureKafkaLag(ctx context.Context, pool *pgxpool.Pool) string {
	var cnt int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM group_members WHERE group_id=1`).Scan(&cnt)
	return fmt.Sprintf("group_members_sample=%d", cnt)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
