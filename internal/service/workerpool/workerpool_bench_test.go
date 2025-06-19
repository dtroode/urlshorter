package workerpool

import (
	"context"
	"sync"
	"testing"
	"time"
)

func BenchmarkPool_Submit(b *testing.B) {
	workerConfigs := []struct {
		workers   int
		queueSize int
		name      string
	}{
		{workers: 10, queueSize: 30, name: "10_workers_30_queue"},
		{workers: 20, queueSize: 60, name: "20_workers_60_queue"},
		{workers: 50, queueSize: 150, name: "50_workers_150_queue"},
		{workers: 100, queueSize: 300, name: "100_workers_300_queue"},
		{workers: 200, queueSize: 600, name: "200_workers_600_queue"},
	}

	taskCount := 50

	for _, config := range workerConfigs {
		b.Run(config.name, func(b *testing.B) {
			pool := NewPool(config.workers, config.queueSize)
			pool.Start()
			defer pool.Close()

			longRunningTask := func(ctx context.Context) (any, error) {
				time.Sleep(300 * time.Millisecond)
				return nil, nil
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				wg.Add(taskCount)

				for j := 0; j < taskCount; j++ {
					go func() {
						defer wg.Done()
						job := pool.Submit(context.Background(), 1*time.Second, longRunningTask, true)
						<-job.ResCh
					}()
				}

				wg.Wait()
			}
		})
	}
}
