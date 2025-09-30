package importer

import (
	"L0/internal/config"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	kaf "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func TestProducer_ProduceMessage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/cp-kafka:7.5.0",
		kafka.WithClusterID("test-cluster"))
	require.NoError(t, err, "failed to start kafka container")

	defer func() {
		if errT := kafkaContainer.Terminate(ctx); errT != nil {
			t.Fatalf("failed to terminate kafka container: %v", errT)
		}
	}()

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err, "failed to get brokers")

	testTopic := "test-topic"
	err = createTopic(brokers[0], testTopic)
	require.NoError(t, err, "failed to create topic")

	cfg := &config.Kafka{
		Brokers: brokers[0],
		Topic:   testTopic,
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	producer := NewProducer(cfg, log)
	testKey := []byte("test-key")
	testValue := []byte("test-value")

	err = producer.ProduceMessage(ctx, testKey, testValue)
	require.NoError(t, err, "failed to produce message")
	err = producer.Close()
	require.NoError(t, err, "failed to close producer")

	reader := kaf.NewReader(kaf.ReaderConfig{
		Brokers:     []string{brokers[0]},
		Topic:       testTopic,
		GroupID:     "test-group",
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     2 * time.Second,
		StartOffset: kaf.FirstOffset,
	})
	defer func() {
		if err = reader.Close(); err != nil {
			t.Fatalf("failed to close reader: %v", err)
		}
	}()

	readCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	msg, err := reader.ReadMessage(readCtx)
	require.NoError(t, err, "failed to read message from reader")

	assert.Equal(t, testKey, msg.Key)
	assert.Equal(t, testValue, msg.Value)
}

func createTopic(brokerAddress string, topic string) error {
	conn, err := kaf.Dial("tcp", brokerAddress)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("failed to close connection: %v", err)
		}
	}()

	topicConfigs := []kaf.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	return conn.CreateTopics(topicConfigs...)
}

type memHandler struct {
	msgs []string
}

func (h *memHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *memHandler) Handle(_ context.Context, r slog.Record) error {
	h.msgs = append(h.msgs, r.Message)
	return nil
}

func (h *memHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *memHandler) WithGroup(_ string) slog.Handler      { return h }

func (h *memHandler) Has(substr string) bool {
	for _, m := range h.msgs {
		if strings.Contains(m, substr) {
			return true
		}
	}
	return false
}

func TestProducer_Completion(t *testing.T) {
	h := &memHandler{}
	log := slog.New(h)
	cfg := &config.Kafka{
		Brokers: "localhost:9092",
		Topic:   "test",
	}
	p := NewProducer(cfg, log)
	defer func() {
		if err := p.Close(); err != nil {
			t.Fatalf("failed to close producer: %v", err)
		}
	}()

	simulatedErr := errors.New("simulated error")
	p.writer.Completion(nil, simulatedErr)

	require.True(t, h.Has("async batch send failed"), "expected log message not found")
}

func TestProducer_ErrorLogger(t *testing.T) {
	h := &memHandler{}
	log := slog.New(h)
	cfg := &config.Kafka{
		Brokers: "localhost:9092",
		Topic:   "test",
	}
	p := NewProducer(cfg, log)
	defer func() {
		if err := p.Close(); err != nil {
			t.Fatalf("failed to close producer: %v", err)
		}
	}()

	p.writer.ErrorLogger.Printf("connection refused to broker %s", "kafka:9092")
	expectedLog := "[KAFKA WRITER ERROR] connection refused to broker kafka:9092"
	require.True(t, h.Has(expectedLog), "expected formatted log was not found")

	brokenLogSubstring := "connection refused to broker %s"
	require.False(t, h.Has(brokenLogSubstring), "log contains unformatted placeholder %s")
}

func TestProducer_WriteMessagesError(t *testing.T) {
	h := &memHandler{}
	log := slog.New(h)
	cfg := &config.Kafka{
		Brokers: "localhost:9092",
		Topic:   "test",
	}
	p := NewProducer(cfg, log)
	err := p.Close()
	require.NoError(t, err, "failed to close producer")

	err = p.ProduceMessage(context.Background(), []byte("test"), []byte("test"))
	require.Error(t, err, "expected error")
}
