package importer

import (
	"context"
	"log/slog"
	"sync"
)

//go:generate go run github.com/vektra/mockery/v2 --name ProducerInter --output ../mocks/importer --case underscore
type ProducerInter interface {
	ProduceMessage(context.Context, []byte, []byte) error
	Close() error
}

//go:generate go run github.com/vektra/mockery/v2 --name GeneratorInter --output ../mocks/importer --case underscore
type GeneratorInter interface {
	GenerateOrder(i int, valid bool) ([]byte, []byte, error)
}

type ImporterService struct {
	producer  ProducerInter
	log       *slog.Logger
	generator GeneratorInter
}

func NewImportService(producer ProducerInter, log *slog.Logger, generator GeneratorInter) *ImporterService {
	return &ImporterService{producer: producer, log: log, generator: generator}
}

func (s *ImporterService) Run(ctx context.Context) error {
	const op = "ImporterService.Run"

	var (
		totalJobs  = 15 //100
		numWorkers = 3  //10
	)
	wg := &sync.WaitGroup{}
	jobs := make(chan int)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go s.worker(ctx, wg, i, jobs)
	}

	s.completeJobs(ctx, jobs, totalJobs)
	close(jobs)
	wg.Wait()
	return nil
}

func (s *ImporterService) worker(ctx context.Context, wg *sync.WaitGroup, workerID int, jobs <-chan int) {
	const op = "ImporterService.worker"

	defer wg.Done()

	for job := range jobs {
		if ctx.Err() != nil {
			s.log.Error("Error generate message", "err", ctx.Err(), "worker", workerID, "job", job, "op", op)
			return
		}
		valid := (job+1)%20 != 0
		key, value, err := s.generator.GenerateOrder(job, valid)
		if err != nil {
			s.log.Error("Error generate message", "err", err, "worker", workerID, "job", job, "op", op)
			continue
		}

		if err = s.producer.ProduceMessage(ctx, key, value); err != nil {
			s.log.Warn("Error adding message to queue", "err", err, "key", key, "worker", workerID, "job", job, "op", op)
		}
	}
}

func (s *ImporterService) completeJobs(ctx context.Context, jobs chan<- int, totalJobs int) {
	const op = "ImporterService.completeJobs"

	for i := 0; i < totalJobs; i++ {
		select {
		case <-ctx.Done():
			s.log.Warn("Context canceled, stopping importer completeJobs", "op", op)
			return
		case jobs <- i:
		}
	}
}

func (s *ImporterService) Close() error {
	const op = "ImporterService.close"
	err := s.producer.Close()
	if err != nil {
		s.log.Error("Error closing producer", "op", op)
		return err
	}
	return nil
}
