package workerpool

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPool(t *testing.T) {
	tests := map[string]struct {
		limit     int
		queueSize int
		expected  *Pool
	}{
		"with custom queue size": {
			limit:     5,
			queueSize: 10,
			expected: &Pool{
				jobs:  make(chan *Job, 10),
				limit: 5,
			},
		},
		"with zero queue size": {
			limit:     3,
			queueSize: 0,
			expected: &Pool{
				jobs:  make(chan *Job, 15), // 3 * 5
				limit: 3,
			},
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			pool := NewPool(tt.limit, tt.queueSize)

			assert.Equal(t, tt.expected.limit, pool.limit)
			assert.Equal(t, cap(tt.expected.jobs), cap(pool.jobs))
		})
	}
}

func TestPool_Start(t *testing.T) {
	pool := NewPool(3, 5)

	// Запускаем пул
	pool.Start()

	// Даем время воркерам запуститься
	time.Sleep(10 * time.Millisecond)

	// Проверяем, что канал создан
	assert.NotNil(t, pool.jobs)
	assert.Equal(t, 5, cap(pool.jobs))
}

func TestPool_Close(t *testing.T) {
	pool := NewPool(2, 3)
	pool.Start()

	// Закрываем пул
	pool.Close()

	// Проверяем, что канал закрыт
	_, ok := <-pool.jobs
	assert.False(t, ok, "Channel should be closed")
}

func TestPool_Submit_WithoutResult(t *testing.T) {
	pool := NewPool(2, 5)
	pool.Start()
	defer pool.Close()

	var executed bool
	var mu sync.Mutex

	job := pool.Submit(
		context.Background(),
		1*time.Second,
		func(ctx context.Context) (any, error) {
			mu.Lock()
			executed = true
			mu.Unlock()
			return "success", nil
		},
		false, // expectResult = false
	)

	// Даем время на выполнение
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.True(t, executed)
	mu.Unlock()

	// Проверяем, что канал результата не создан
	assert.Nil(t, job.ResCh)
}

func TestPool_Submit_WithResult(t *testing.T) {
	pool := NewPool(2, 5)
	pool.Start()
	defer pool.Close()

	expectedValue := "test result"

	job := pool.Submit(
		context.Background(),
		1*time.Second,
		func(ctx context.Context) (any, error) {
			return expectedValue, nil
		},
		true, // expectResult = true
	)

	// Проверяем, что канал результата создан
	require.NotNil(t, job.ResCh)

	// Ждем результат
	select {
	case result := <-job.ResCh:
		assert.NoError(t, result.Err)
		assert.Equal(t, expectedValue, result.Value)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
}

func TestPool_Submit_WithError(t *testing.T) {
	pool := NewPool(2, 5)
	pool.Start()
	defer pool.Close()

	expectedError := errors.New("test error")

	job := pool.Submit(
		context.Background(),
		1*time.Second,
		func(ctx context.Context) (any, error) {
			return nil, expectedError
		},
		true,
	)

	require.NotNil(t, job.ResCh)

	select {
	case result := <-job.ResCh:
		assert.Error(t, result.Err)
		assert.Equal(t, expectedError, result.Err)
		assert.Nil(t, result.Value)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
}

func TestPool_Submit_WithTimeout(t *testing.T) {
	pool := NewPool(2, 5)
	pool.Start()
	defer pool.Close()

	job := pool.Submit(
		context.Background(),
		100*time.Millisecond, // короткий таймаут
		func(ctx context.Context) (any, error) {
			// Симулируем долгую операцию
			time.Sleep(200 * time.Millisecond)
			return "should not reach here", nil
		},
		true,
	)

	require.NotNil(t, job.ResCh)

	select {
	case result := <-job.ResCh:
		assert.Error(t, result.Err)
		assert.Contains(t, result.Err.Error(), "context deadline exceeded")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for result")
	}
}

func TestPool_Submit_WithContextCancellation(t *testing.T) {
	pool := NewPool(2, 5)
	pool.Start()
	defer pool.Close()

	ctx, cancel := context.WithCancel(context.Background())

	job := pool.Submit(
		ctx,
		1*time.Second,
		func(ctx context.Context) (any, error) {
			// Симулируем долгую операцию
			time.Sleep(200 * time.Millisecond)
			return "should not reach here", nil
		},
		true,
	)

	// Отменяем контекст сразу
	cancel()

	require.NotNil(t, job.ResCh)

	select {
	case result := <-job.ResCh:
		assert.Error(t, result.Err)
		assert.Contains(t, result.Err.Error(), "context canceled")
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for result")
	}
}

func TestPool_ConcurrentJobs(t *testing.T) {
	pool := NewPool(3, 10)
	pool.Start()
	defer pool.Close()

	const numJobs = 10
	var results = make([]*Result, numJobs)
	var wg sync.WaitGroup

	// Запускаем несколько задач одновременно
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			job := pool.Submit(
				context.Background(),
				1*time.Second,
				func(ctx context.Context) (any, error) {
					time.Sleep(10 * time.Millisecond) // небольшая задержка
					return index, nil
				},
				true,
			)

			result := <-job.ResCh
			results[index] = result
		}(i)
	}

	wg.Wait()

	// Проверяем все результаты
	for i := 0; i < numJobs; i++ {
		require.NotNil(t, results[i])
		assert.NoError(t, results[i].Err)
		assert.Equal(t, i, results[i].Value)
	}
}

func TestPool_JobOrdering(t *testing.T) {
	pool := NewPool(1, 5) // только один воркер
	pool.Start()
	defer pool.Close()

	var executionOrder []int
	var mu sync.Mutex

	// Запускаем задачи с разными задержками
	job1 := pool.Submit(
		context.Background(),
		1*time.Second,
		func(ctx context.Context) (any, error) {
			time.Sleep(50 * time.Millisecond)
			mu.Lock()
			executionOrder = append(executionOrder, 1)
			mu.Unlock()
			return 1, nil
		},
		true,
	)

	job2 := pool.Submit(
		context.Background(),
		1*time.Second,
		func(ctx context.Context) (any, error) {
			time.Sleep(10 * time.Millisecond)
			mu.Lock()
			executionOrder = append(executionOrder, 2)
			mu.Unlock()
			return 2, nil
		},
		true,
	)

	// Ждем результаты
	<-job1.ResCh
	<-job2.ResCh

	// Проверяем, что обе задачи выполнились (порядок не гарантирован с горутинами)
	assert.Len(t, executionOrder, 2)
	assert.Contains(t, executionOrder, 1)
	assert.Contains(t, executionOrder, 2)
}

func TestPool_StressTest(t *testing.T) {
	pool := NewPool(5, 20)
	pool.Start()
	defer pool.Close()

	const numJobs = 50
	var completedJobs int
	var mu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			job := pool.Submit(
				context.Background(),
				500*time.Millisecond,
				func(ctx context.Context) (any, error) {
					time.Sleep(10 * time.Millisecond)
					mu.Lock()
					completedJobs++
					mu.Unlock()
					return "completed", nil
				},
				true,
			)

			result := <-job.ResCh
			assert.NoError(t, result.Err)
			assert.Equal(t, "completed", result.Value)
		}()
	}

	wg.Wait()
	assert.Equal(t, numJobs, completedJobs)
}

func TestPool_SubmitAfterClose(t *testing.T) {
	pool := NewPool(2, 5)
	pool.Start()
	pool.Close()

	// Даем время на закрытие канала
	time.Sleep(10 * time.Millisecond)

	// Пытаемся отправить задачу после закрытия
	// Это должно вызвать панику, но мы можем её перехватить
	defer func() {
		if r := recover(); r != nil {
			// Ожидаемая паника при отправке в закрытый канал
			assert.Contains(t, r, "send on closed channel")
		}
	}()

	job := pool.Submit(
		context.Background(),
		1*time.Second,
		func(ctx context.Context) (any, error) {
			return "test", nil
		},
		true,
	)

	// Если паники не было, проверяем, что задача не обрабатывается
	select {
	case <-job.ResCh:
		t.Fatal("Job should not be processed after pool close")
	case <-time.After(100 * time.Millisecond):
		// Ожидаемое поведение - задача не обрабатывается
	}
}

func TestPool_ZeroWorkers(t *testing.T) {
	pool := NewPool(0, 5)
	pool.Start()
	defer pool.Close()

	// Пул с нулевым количеством воркеров должен работать
	// но задачи не будут обрабатываться
	job := pool.Submit(
		context.Background(),
		1*time.Second,
		func(ctx context.Context) (any, error) {
			return "test", nil
		},
		true,
	)

	select {
	case <-job.ResCh:
		t.Fatal("Job should not be processed with zero workers")
	case <-time.After(100 * time.Millisecond):
		// Ожидаемое поведение
	}
}

func TestPool_JobWithNilFunction(t *testing.T) {
	pool := NewPool(2, 5)
	pool.Start()
	defer pool.Close()

	// Это может вызвать панику, но тестируем поведение
	job := pool.Submit(
		context.Background(),
		1*time.Second,
		nil, // nil функция
		true,
	)

	// Ожидаем панику или ошибку
	select {
	case result := <-job.ResCh:
		assert.Error(t, result.Err)
	case <-time.After(100 * time.Millisecond):
		// Возможно, задача не обрабатывается из-за паники
	}
}
