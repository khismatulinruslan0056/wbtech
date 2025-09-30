package importer

import (
	"L0/internal/mocks/importer"
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestImporterService_Run(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success generate and produce all jobs", func(t *testing.T) {
		mockProducer := new(mocks.ProducerInter)
		mockGenerator := new(mocks.GeneratorInter)

		service := NewImportService(mockProducer, log, mockGenerator)

		testKey := []byte("testKey")
		testValue := []byte("testValue")
		totalJobs := 100
		mockGenerator.On("GenerateOrder", mock.AnythingOfType("int"), mock.AnythingOfType("bool")).
			Return(testKey, testValue, nil).Times(totalJobs)
		mockProducer.On("ProduceMessage", mock.Anything, testKey, testValue).
			Return(nil).Times(totalJobs)

		err := service.Run(context.Background())

		assert.NoError(t, err)

		mockGenerator.AssertCalled(t, "GenerateOrder", mock.AnythingOfType("int"), mock.AnythingOfType("bool"))
		mockProducer.AssertCalled(t, "ProduceMessage", mock.Anything, testKey, testValue)
	})

	t.Run("not call producer when generator returns an error", func(t *testing.T) {
		mockProducer := new(mocks.ProducerInter)
		mockGenerator := new(mocks.GeneratorInter)

		service := NewImportService(mockProducer, log, mockGenerator)

		generatorError := errors.New("generator error")
		mockGenerator.On("GenerateOrder", mock.AnythingOfType("int"), mock.AnythingOfType("bool")).
			Return(nil, nil, generatorError)

		err := service.Run(context.Background())

		assert.NoError(t, err)

		mockGenerator.AssertCalled(t, "GenerateOrder", mock.AnythingOfType("int"), mock.AnythingOfType("bool"))
		mockProducer.AssertNotCalled(t, "ProduceMessage", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("stop sending when context is cancelled", func(t *testing.T) {
		mockProducer := new(mocks.ProducerInter)
		mockGenerator := new(mocks.GeneratorInter)

		service := NewImportService(mockProducer, log, mockGenerator)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		mockGenerator.On("GenerateOrder", mock.AnythingOfType("int"), mock.AnythingOfType("bool")).
			Return(nil, nil, nil)
		mockProducer.On("ProduceMessage", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := service.Run(ctx)
		assert.NoError(t, err)

		assert.Less(t, len(mockGenerator.Calls), 100, "generator must be called less than 100 times")
	})

	t.Run("error from producer", func(t *testing.T) {
		mockProducer := new(mocks.ProducerInter)
		mockGenerator := new(mocks.GeneratorInter)

		service := NewImportService(mockProducer, log, mockGenerator)

		testKey := []byte("testKey")
		testValue := []byte("testValue")
		totalJobs := 100
		producerError := errors.New("producer error")

		mockGenerator.On("GenerateOrder", mock.AnythingOfType("int"), mock.AnythingOfType("bool")).
			Return(testKey, testValue, nil).Times(totalJobs)
		mockProducer.On("ProduceMessage", mock.Anything, testKey, testValue).
			Return(producerError).Times(totalJobs)

		err := service.Run(context.Background())

		assert.NoError(t, err)

		mockGenerator.AssertCalled(t, "GenerateOrder", mock.AnythingOfType("int"), mock.AnythingOfType("bool"))
		mockProducer.AssertCalled(t, "ProduceMessage", mock.Anything, testKey, testValue)
	})
}

func TestImporterService_Close(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	t.Run("correct graceful shutdown", func(t *testing.T) {
		mockProducer := new(mocks.ProducerInter)
		mockGenerator := new(mocks.GeneratorInter)
		service := NewImportService(mockProducer, log, mockGenerator)

		mockProducer.On("Close").Return(nil)

		err := service.Close()

		assert.NoError(t, err)
	})
	t.Run("error graceful shutdown", func(t *testing.T) {
		mockProducer := new(mocks.ProducerInter)
		mockGenerator := new(mocks.GeneratorInter)
		producerError := errors.New("producer error")
		service := NewImportService(mockProducer, log, mockGenerator)

		mockProducer.On("Close").Return(producerError)

		err := service.Close()

		assert.Error(t, err)
	})
}

func TestWorker_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	wg := &sync.WaitGroup{}
	jobs := make(chan int, 1)
	jobs <- 1
	close(jobs)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockProducer := new(mocks.ProducerInter)
	mockGenerator := new(mocks.GeneratorInter)

	s := &ImporterService{
		producer:  mockProducer,
		generator: mockGenerator,
		log:       log,
	}

	wg.Add(1)
	go s.worker(ctx, wg, 1, jobs)
	wg.Wait()
}

func TestCompleteJobs_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	jobs := make(chan int, 1)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := &ImporterService{log: log}

	s.completeJobs(ctx, jobs, 10)
	close(jobs)
}
