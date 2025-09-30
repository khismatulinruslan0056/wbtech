package dlq

import (
	"L0/internal/config"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(cfg *config.Kafka) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(cfg.Brokers, ",")...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	return &Producer{writer: w}
}

func (p *Producer) Produce(ctx context.Context, key, value []byte) error {
	const op = "Producer.Produce"
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type DLQMessage struct {
	Value            []byte    `json:"value"`
	Error            string    `json:"error"`
	FailureTimestamp time.Time `json:"failure_timestamp"`
	Topic            string    `json:"topic"`
	Partition        int       `json:"partition"`
	Offset           int64     `json:"offset"`
}
