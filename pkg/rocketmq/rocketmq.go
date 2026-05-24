package rocketmq

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

// Producer 向指定 Topic + Tag 发布 JSON 消息。
type Producer struct {
	p rocketmq.Producer
}

// resolveNameServers 将 host:port 中的主机名解析为 IP。
// rocketmq-client-go 仅接受 IP（见 primitive.verifyIP），K8s 配置常用 Service DNS 名。
func resolveNameServers(addrs []string) ([]string, error) {
	out := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("nameserver %q: %w", addr, err)
		}
		if ip := net.ParseIP(host); ip != nil {
			out = append(out, net.JoinHostPort(ip.String(), port))
			continue
		}
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, fmt.Errorf("nameserver %q: lookup: %w", addr, err)
		}
		var picked net.IP
		for _, ip := range ips {
			if v4 := ip.To4(); v4 != nil {
				picked = v4
				break
			}
		}
		if picked == nil {
			picked = ips[0]
		}
		out = append(out, net.JoinHostPort(picked.String(), port))
	}
	return out, nil
}

func NewProducer(nameServers []string) (*Producer, error) {
	resolved, err := resolveNameServers(nameServers)
	if err != nil {
		return nil, err
	}
	p, err := rocketmq.NewProducer(
		producer.WithNameServer(resolved),
		producer.WithRetry(2),
	)
	if err != nil {
		return nil, err
	}
	if err := p.Start(); err != nil {
		return nil, err
	}
	return &Producer{p: p}, nil
}

func (p *Producer) Close() error {
	if p == nil || p.p == nil {
		return nil
	}
	return p.p.Shutdown()
}

func (p *Producer) PublishJSON(ctx context.Context, topic, tag, key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	msg := primitive.NewMessage(topic, b)
	msg.WithTag(tag)
	if key != "" {
		msg.WithKeys([]string{key})
	}
	_, err = p.p.SendSync(ctx, msg)
	return err
}

// ConsumerConfig Push 消费配置；Tag 为 RocketMQ Tag 过滤表达式（单 Tag 或 `a || b`）。
type ConsumerConfig struct {
	NameServers []string
	Topic       string
	Group       string
	Tag         string
	Broadcast   bool
}

// RunPushConsumer 阻塞消费直到 ctx 取消。
func RunPushConsumer(ctx context.Context, cfg ConsumerConfig, handle func(ctx context.Context, body []byte) error) error {
	resolved, err := resolveNameServers(cfg.NameServers)
	if err != nil {
		return err
	}
	opts := []consumer.Option{
		consumer.WithNameServer(resolved),
		consumer.WithGroupName(cfg.Group),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
	}
	if cfg.Broadcast {
		opts = append(opts, consumer.WithConsumerModel(consumer.BroadCasting))
	}
	c, err := rocketmq.NewPushConsumer(opts...)
	if err != nil {
		return err
	}
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: cfg.Tag,
	}
	if err := c.Subscribe(cfg.Topic, selector, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, m := range msgs {
			if err := handle(ctx, m.Body); err != nil {
				return consumer.ConsumeRetryLater, nil
			}
		}
		return consumer.ConsumeSuccess, nil
	}); err != nil {
		return err
	}
	if err := c.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		_ = c.Shutdown()
	}()
	wg.Wait()
	return ctx.Err()
}

// MustProducer 启动失败时 panic（与 go-zero Must 风格一致）。
func MustProducer(nameServers []string) *Producer {
	p, err := NewProducer(nameServers)
	if err != nil {
		panic(fmt.Sprintf("rocketmq producer: %v", err))
	}
	return p
}
