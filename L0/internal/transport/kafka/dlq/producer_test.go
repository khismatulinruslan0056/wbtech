package dlq

import (
	"L0/internal/config"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kafkaTC "github.com/testcontainers/testcontainers-go/modules/kafka"
)

func TestProducer_Produce(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()

	kafkaContainer, err := kafkaTC.Run(ctx,
		"confluentinc/cp-kafka:7.5.0",
		kafkaTC.WithClusterID("test-cluster"))
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

	producer := NewProducer(cfg)
	testKey := []byte("test-key")
	testValue := []byte("test-value")

	err = producer.Produce(ctx, testKey, testValue)
	require.NoError(t, err, "failed to produce message")
	err = producer.Close()
	require.NoError(t, err, "failed to close producer")

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{brokers[0]},
		Topic:       testTopic,
		GroupID:     "test-group",
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     2 * time.Second,
		StartOffset: kafka.FirstOffset,
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
	conn, err := kafka.Dial("tcp", brokerAddress)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("failed to close connection: %v", err)
		}
	}()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	return conn.CreateTopics(topicConfigs...)
}

func TestProducerErr(t *testing.T) {
	cfg := &config.Kafka{
		Brokers: "localhost:9092",
		Topic:   "test",
	}
	p := NewProducer(cfg)
	err := p.Close()
	require.NoError(t, err, "failed to close producer")

	err = p.Produce(context.Background(), []byte("test"), []byte("test"))
	require.Error(t, err, "expected error")
}
