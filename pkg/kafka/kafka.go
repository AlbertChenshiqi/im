package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

func NewWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
	}
}

func NewReader(brokers []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})
}

func PublishJSON(ctx context.Context, w *kafka.Writer, key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: b,
	})
}

func ReadJSON(ctx context.Context, r *kafka.Reader, v any) error {
	m, err := r.FetchMessage(ctx)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(m.Value, v); err != nil {
		_ = r.CommitMessages(ctx, m)
		return err
	}
	return r.CommitMessages(ctx, m)
}

func ReadJSONMessage(ctx context.Context, r *kafka.Reader, v any) (kafka.Message, error) {
	m, err := r.FetchMessage(ctx)
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(m.Value, v); err != nil {
		return m, err
	}
	return m, nil
}
