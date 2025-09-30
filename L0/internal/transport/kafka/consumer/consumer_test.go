package consumer

import (
	"L0/internal/config"
	mocks "L0/internal/mocks/transport/kafka/consumer"
	k "L0/internal/transport/kafka"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	kafkaTC "github.com/testcontainers/testcontainers-go/modules/kafka"
)

type MockReader struct {
	mock.Mock
}

func (m *MockReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	args := m.Called(ctx)
	if msg, ok := args.Get(0).(kafka.Message); ok {
		return msg, args.Error(1)
	}
	return kafka.Message{}, args.Error(1)
}

func (m *MockReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	allArgs := []interface{}{ctx}
	for _, msg := range msgs {
		allArgs = append(allArgs, msg)
	}
	args := m.Called(allArgs...)
	return args.Error(0)
}

func (m *MockReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestWorker(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()
	testMsg := kafka.Message{
		Key:   []byte("test-key"),
		Value: []byte("test-value"),
	}

	t.Run("success", func(t *testing.T) {
		mockReader := new(MockReader)
		mockHandler := new(mocks.Handler)
		mockDLQProducer := new(mocks.DLQProducer)

		ch := make(chan kafka.Message)
		consumer := &Consumer{
			reader:      mockReader,
			handler:     mockHandler,
			log:         log,
			numWorkers:  1,
			wg:          sync.WaitGroup{},
			dLQProducer: mockDLQProducer,
		}

		mockHandler.On("Handle", mock.Anything, testMsg).Return(nil)
		consumer.wg.Add(1)
		go func() {
			consumer.worker(ctx, 1, ch)
		}()

		ch <- testMsg
		close(ch)
		consumer.wg.Wait()
		mockHandler.AssertExpectations(t)
		mockDLQProducer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything)
		mockReader.AssertNotCalled(t, "CommitMessages", mock.Anything, mock.Anything)
	})

	t.Run("unsuccess, dlq/commit success", func(t *testing.T) {
		mockReader := new(MockReader)
		mockHandler := new(mocks.Handler)
		mockDLQProducer := new(mocks.DLQProducer)

		ch := make(chan kafka.Message)
		consumer := &Consumer{
			reader:      mockReader,
			handler:     mockHandler,
			log:         log,
			numWorkers:  1,
			wg:          sync.WaitGroup{},
			dLQProducer: mockDLQProducer,
		}
		handleErr := errors.New("handle error")
		mockHandler.On("Handle", mock.Anything, testMsg).Return(handleErr)
		mockDLQProducer.On("Produce", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		mockReader.On("CommitMessages", mock.Anything, mock.Anything).Return(nil).Once()
		consumer.wg.Add(1)
		go func() {
			consumer.worker(ctx, 1, ch)
		}()

		ch <- testMsg
		close(ch)
		consumer.wg.Wait()
		mockHandler.AssertExpectations(t)
		mockDLQProducer.AssertCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything)
		mockReader.AssertCalled(t, "CommitMessages", mock.Anything, mock.Anything)
	})
	t.Run("unsuccess, dlq/commit unsuccess", func(t *testing.T) {
		mockReader := new(MockReader)
		mockHandler := new(mocks.Handler)
		mockDLQProducer := new(mocks.DLQProducer)

		ch := make(chan kafka.Message)
		consumer := &Consumer{
			reader:      mockReader,
			handler:     mockHandler,
			log:         log,
			numWorkers:  1,
			wg:          sync.WaitGroup{},
			dLQProducer: mockDLQProducer,
		}
		var (
			handleErr = errors.New("handle error")
			retryErr  = errors.New("retry error")
			commitErr = errors.New("commit error")
		)
		mockHandler.On("Handle", mock.Anything, testMsg).Return(handleErr)
		mockDLQProducer.On("Produce", mock.Anything, mock.Anything, mock.Anything).Return(retryErr).Once()
		mockReader.On("CommitMessages", mock.Anything, mock.Anything).Return(commitErr).Once()
		consumer.wg.Add(1)
		go func() {
			consumer.worker(ctx, 1, ch)
		}()

		ch <- testMsg
		close(ch)
		consumer.wg.Wait()

		mockHandler.AssertExpectations(t)
		mockDLQProducer.AssertCalled(t, "Produce", mock.Anything, mock.Anything, mock.Anything)
		mockReader.AssertCalled(t, "CommitMessages", mock.Anything, mock.Anything)
	})

}

func TestProcessMsgWithRetries(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	ctx := context.Background()
	msg := &kafka.Message{
		Key:   []byte("test"),
		Value: []byte("test"),
	}

	cfg := &config.Kafka{
		Brokers: "localhost:9092",
		Topic:   "test",
	}
	concur := 0
	mockHandler := new(mocks.Handler)
	consumer := NewConsumer(cfg, mockHandler, concur, log)
	t.Run("success", func(t *testing.T) {
		mockHandler.On("Handle", mock.Anything, mock.Anything).Return(nil).Once()
		err := consumer.processMsgWithRetries(ctx, *msg)
		require.NoError(t, err, "expected no error")
	})
	t.Run("success retry", func(t *testing.T) {
		retryErr := errors.New("retry error")
		mockHandler.On("Handle", mock.Anything, mock.Anything).Return(retryErr).Twice()
		mockHandler.On("Handle", mock.Anything, mock.Anything).Return(nil).Once()
		err := consumer.processMsgWithRetries(ctx, *msg)
		require.NoError(t, err, "expected no error")
	})
	t.Run("unsuccess", func(t *testing.T) {
		retryErr := errors.New("retry error")
		mockHandler.On("Handle", mock.Anything, mock.Anything).Return(retryErr)
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err := consumer.processMsgWithRetries(ctx, *msg)
		require.Error(t, err, "expected error")
		require.ErrorIs(t, err, retryErr, "expected wrap error")
		mockHandler.AssertCalled(t, "Handle", mock.Anything, *msg)
	})
	t.Run("unsuccess noretry", func(t *testing.T) {
		retryErr := errors.New("retry error")
		mockHandler.On("Handle", mock.Anything, mock.Anything).Return(k.NewNonRetriableError(retryErr)).Once()
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		err := consumer.processMsgWithRetries(ctx, *msg)
		require.Error(t, err, "expected error")
		var errN k.NonRetriableError
		require.ErrorAs(t, err, &errN, "expected wrap error")
		mockHandler.AssertCalled(t, "Handle", mock.Anything, *msg)
	})
}

func TestConsumer_Close(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("successful close", func(t *testing.T) {
		mockReader := new(MockReader)
		mockReader.On("Close").Return(nil).Once()

		consumer := &Consumer{
			reader: mockReader,
			log:    log,
		}

		err := consumer.Close()

		require.NoError(t, err)
		mockReader.AssertExpectations(t)
	})

	t.Run("close returns error", func(t *testing.T) {
		mockReader := new(MockReader)
		expectedErr := errors.New("failed to close connection")
		mockReader.On("Close").Return(expectedErr).Once()

		consumer := &Consumer{
			reader: mockReader,
			log:    log,
		}

		err := consumer.Close()

		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
		mockReader.AssertExpectations(t)
	})
}

func TestConsumer_RunIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.Background()
	kafkaContainer, err := kafkaTC.Run(ctx,
		"confluentinc/cp-kafka:7.5.0",
		kafkaTC.WithClusterID("test-cluster"))

	require.NoError(t, err, "failed to start kafka container")

	defer func() {
		require.NoError(t, kafkaContainer.Terminate(ctx), "failed to terminate kafka container")
	}()

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err, "failed to get brokers")
	testTopic := "test-consumer-topic"
	err = createTopic(brokers[0], testTopic)
	require.NoError(t, err, "failed to create topic")

	cfg := &config.Kafka{
		Brokers: brokers[0],
		Topic:   testTopic,
		GroupID: "test-group-run",
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	testKey := []byte("test-key")
	testValue := []byte("test-value")

	messageHandled := make(chan struct{})

	mockHandler := new(mocks.Handler)
	mockHandler.On("Handle", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			msg := args.Get(1).(kafka.Message)
			assert.Equal(t, testKey, msg.Key)
			assert.Equal(t, testValue, msg.Value)
			close(messageHandled)

		}).Return(nil).Once()

	consumer := NewConsumer(cfg, mockHandler, 5, log)
	cfgReader := kafka.ReaderConfig{
		Brokers:        strings.Split(cfg.Brokers, ", "),
		GroupID:        cfg.GroupID,
		Topic:          cfg.Topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
		StartOffset:    kafka.FirstOffset,
	}
	consumer.reader = kafka.NewReader(cfgReader)
	runCtx, cancelRun := context.WithCancel(context.Background())

	go func() {
		consumer.Run(runCtx)
	}()

	w := &kafka.Writer{
		Addr:  kafka.TCP(strings.Split(cfg.Brokers, ",")...),
		Topic: cfg.Topic,
	}

	err = w.WriteMessages(ctx, kafka.Message{
		Key:   testKey,
		Value: testValue,
	})
	require.NoError(t, err, "failed to write message")
	require.NoError(t, w.Close(), "failed to close writer")
	time.Sleep(1000 * time.Millisecond)
	select {
	case <-messageHandled:
	case <-time.After(60 * time.Second):
		t.Fatal("failed to receive message after 60 seconds")
	}
	cancelRun()
	err = consumer.Close()
	require.NoError(t, err, "failed to close consumer")

	mockHandler.AssertExpectations(t)
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

func TestConsumer_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockReader := new(MockReader)

	mockReader.On("ReadMessage", mock.Anything).Return(kafka.Message{}, errors.New("read error")).Once()

	mockReader.On("ReadMessage", mock.Anything).Return(kafka.Message{}, context.Canceled)

	mockHandler := new(mocks.Handler)

	consumer := &Consumer{
		reader:     mockReader,
		handler:    mockHandler,
		numWorkers: 1,
		log:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	done := make(chan struct{})
	go func() {
		consumer.Run(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Consumer.Run did not exit after context cancel")
	}

	mockReader.AssertExpectations(t)

}
