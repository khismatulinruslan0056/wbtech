package consumer

import (
	"L0/internal/config"
	k "L0/internal/transport/kafka"
	"L0/internal/transport/kafka/dlq"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/segmentio/kafka-go"
)

//go:generate go run github.com/vektra/mockery/v2 --name Handler --output ../../../mocks/transport/kafka/consumer --case underscore
type Handler interface {
	Handle(ctx context.Context, msg kafka.Message) error
}

//go:generate go run github.com/vektra/mockery/v2 --name DLQProducer --output ../../../mocks/transport/kafka/consumer --case underscore
type DLQProducer interface {
	Produce(ctx context.Context, key, value []byte) error
}

//go:generate go run github.com/vektra/mockery/v2 --name Reader --output ../../../mocks/transport/kafka/consumer --case underscore
type Reader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Consumer struct {
	reader      Reader
	handler     Handler
	log         *slog.Logger
	numWorkers  int
	wg          sync.WaitGroup
	dLQProducer DLQProducer
}

func NewConsumer(cfg *config.Kafka, handler Handler, numWorkers int, log *slog.Logger) *Consumer {
	if numWorkers < 1 {
		numWorkers = 1
	}
	slogKafkaErrorLogger := kafka.LoggerFunc(func(message string, args ...interface{}) {
		formattedMessage := fmt.Sprintf(message, args...)
		log.Error("[KAFKA WRITER ERROR] " + formattedMessage)
	})

	dLQProducer := dlq.NewProducer(cfg)
	cfgReader := kafka.ReaderConfig{
		Brokers:        strings.Split(cfg.Brokers, ", "),
		GroupID:        cfg.GroupID,
		Topic:          cfg.Topic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 0,
		StartOffset:    kafka.LastOffset,
		ErrorLogger:    slogKafkaErrorLogger,
	}

	return &Consumer{
		reader:      kafka.NewReader(cfgReader),
		numWorkers:  numWorkers,
		log:         log,
		handler:     handler,
		dLQProducer: dLQProducer,
	}
}

func (c *Consumer) Run(ctx context.Context) {

	messages := make(chan kafka.Message, c.numWorkers)

	c.wg.Add(c.numWorkers)

	for i := 0; i < c.numWorkers; i++ {
		go c.worker(ctx, i, messages)
	}

	readBo := backoff.NewExponentialBackOff()
	readBo.InitialInterval = 500 * time.Millisecond
	readBo.Multiplier = 1.7
	readBo.MaxInterval = 5 * time.Second
	readBo.MaxElapsedTime = 0

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				break
			}
			if isTransientKafkaErr(err) {
				wait := readBo.NextBackOff()
				select {
				case <-time.After(wait):
					continue
				case <-ctx.Done():
					break
				}
			}

			time.Sleep(time.Second)
			continue
		}

		readBo.Reset()

		select {
		case messages <- msg:
		case <-ctx.Done():
			break
		}
	}
	close(messages)
	c.wg.Wait()
}

func (c *Consumer) worker(ctx context.Context, id int, messages <-chan kafka.Message) {
	defer c.wg.Done()
	for msg := range messages {
		if err := c.processMsgWithRetries(ctx, msg); err != nil {
			c.log.Error("All attempts to send a message failed",
				"worker", id, "key", string(msg.Key), "err", err)

			dlqValue, err := c.enrichMsgForDLQ(msg, err)
			if err != nil {
				c.log.Warn("dlqValue is empty, message may be lost",
					"worker", id, "key", string(msg.Key))
			}
			if dlqValue != nil {
				if dlqErr := c.dLQProducer.Produce(ctx, msg.Key, dlqValue); dlqErr != nil {
					c.log.Warn("Failed to send to dlq, message may be lost",
						"worker", id, "key", string(msg.Key), "err", dlqErr)
				} else {
					c.log.Info("Message sent to dlq",
						"worker", id, "key", string(msg.Key), "err", dlqErr)
				}
			}

			if err = c.reader.CommitMessages(ctx, msg); err != nil {
				c.log.Error("Offset commit error",
					"worker", id, "key", string(msg.Key), "err", err)
			}
		}
	}
}

func (c *Consumer) processMsgWithRetries(ctx context.Context, msg kafka.Message) error {
	const op = "consumer.processMsgWithRetries"
	handleCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := c.handler.Handle(handleCtx, msg)
	if err == nil {
		return nil
	}
	var nonRetriableErr k.NonRetriableError
	if errors.As(err, &nonRetriableErr) {
		c.log.Warn("Non-retriable error, sending to DLQ immediately",
			"key", string(msg.Key),
			"error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = 30 * time.Second

	operation := func() error {
		handleCtx, cancel = context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		return c.handler.Handle(handleCtx, msg)
	}

	if err = backoff.Retry(operation, expBackoff); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Consumer) enrichMsgForDLQ(msg kafka.Message, err error) ([]byte, error) {
	const op = "Consumer.enrichMsgForDLQ"
	msgDLQ := &dlq.DLQMessage{
		Value:            msg.Value,
		Error:            err.Error(),
		FailureTimestamp: time.Now().UTC(),
		Topic:            msg.Topic,
		Partition:        msg.Partition,
		Offset:           msg.Offset,
	}

	msgRes, err := json.Marshal(msgDLQ)
	if err != nil {
		c.log.Error("error marshalling msgDLQ", "err", err, "op", op)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return msgRes, nil
}

func (c *Consumer) Close() error {
	const op = "consumer.Close"
	if err := c.reader.Close(); err != nil {
		c.log.Error("error closing reader", "err", err)
		return err
	}
	return nil
}

func isTransientKafkaErr(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "coordinator not available") ||
		strings.Contains(s, "not coordinator") ||
		strings.Contains(s, "group coordinator not available") ||
		strings.Contains(s, "group load in progress") ||
		strings.Contains(s, "rebalancing") ||
		strings.Contains(s, "leader not available") ||
		strings.Contains(s, "unknown topic or partition") ||
		strings.Contains(s, "broker not available") ||
		strings.Contains(s, "connection refused") ||
		strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "transport is closing")
}
