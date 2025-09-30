package importer

import (
	"L0/internal/config"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewProducer(cfg *config.Kafka, log *slog.Logger) *Producer {
	slogKafkaErrorLogger := kafka.LoggerFunc(func(message string, args ...interface{}) {
		formattedMessage := fmt.Sprintf(message, args...)
		log.Error("[KAFKA WRITER ERROR] " + formattedMessage)
	})

	w := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(cfg.Brokers, ",")...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  10,
		BatchSize:    1000,
		BatchTimeout: 10 * time.Millisecond,
		Compression:  kafka.Snappy,
		Async:        true,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				log.Error("async batch send failed", slog.String("err", err.Error()))
			}
		},
		AllowAutoTopicCreation: true,
		ErrorLogger:            slogKafkaErrorLogger,
	}

	return &Producer{w, log}
}

func (p *Producer) ProduceMessage(ctx context.Context, key []byte, value []byte) error {
	const op = "Producer.ProduceMessage"
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.log.Error("failed to put message into queue")
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
