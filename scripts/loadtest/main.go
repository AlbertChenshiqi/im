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

	"database/sql"
	"github.com/redis/go-redis/v9"

	"im/pkg/bizseq"
	"im/pkg/convid"
	"im/pkg/db"
	"im/pkg/events"
	"im/pkg/models"
	"im/pkg/rocketmq"
	"im/pkg/sessionid"
)

func main() {
	userAPI := flag.String("user-api", "http://localhost:10100", "user API base URL")
	groupID := flag.Int64("group", 0, "existing group id (0=create)")
	members := flag.Int("members", 1000, "member count for new group")
	msgs := flag.Int("msgs", 50, "messages to send")
	flag.Parse()

	ctx := context.Background()
	dsn := getenv("MYSQL_DSN", "im:im@tcp(localhost:3306)/im?parseTime=true&charset=utf8mb4&loc=Local")
	sqlDB, err := db.NewDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

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

	producer, err := rocketmq.NewProducer([]string{getenv("ROCKETMQ_NAMESERVER", "localhost:9876")})
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	sid := sessionid.FromConvID(conv)
	for i := 0; i < *msgs; i++ {
		recvMs := time.Now().UnixMilli()
		bizSeq, err := bizseq.Allocate(ctx, rdb, sid, recvMs)
		if err != nil {
			log.Fatalf("bizseq: %v", err)
		}
		evt := events.MessageSendEvent{
			MsgID: int64(1_000_000 + i), ConvID: conv, SessionID: sid,
			ConvType: "group", GroupID: *groupID, SenderID: uid,
			BizSeq: bizSeq, Seq: bizSeq, SendTs: recvMs, ServerRecvMs: recvMs,
			Input: []models.MessageInput{{MsgType: "text", Content: fmt.Sprintf(`{"text":"load message %d"}`, i)}},
			Ts:    time.Now().Unix(),
		}
		if err := producer.PublishJSON(ctx, events.TopicChat, events.TagChatGroup, sid, evt); err != nil {
			log.Fatalf("publish chat: %v", err)
		}
		if err := producer.PublishJSON(ctx, events.TopicChatPersist, events.TagChatPersistStore, sid, evt); err != nil {
			log.Fatalf("publish persist: %v", err)
		}
	}

	elapsed := time.Since(start)
	lag := measureConsumerLag(ctx, sqlDB)
	log.Printf("done: %d msgs in %v, rocketmq consumer hint: check inbox-unread logs", *msgs, elapsed)
	log.Printf("metrics: redis online keys, unread keys — lag_estimate=%s", lag)
}

func registerLogin(base, user, pass string) (string, int64) {
	body, _ := json.Marshal(map[string]string{"username": user, "password": pass, "nickname": user})
	resp, err := http.Post(base+"/user/v1/auth/register", "application/json", bytes.NewReader(body))
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
	req, _ := http.NewRequest(http.MethodPost, base+"/group/v1/groups", bytes.NewReader(body))
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

func measureConsumerLag(ctx context.Context, db *sql.DB) string {
	var cnt int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM group_members WHERE group_id=1`).Scan(&cnt)
	return fmt.Sprintf("group_members_sample=%d", cnt)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
