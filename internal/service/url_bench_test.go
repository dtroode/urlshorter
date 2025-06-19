package service

import (
	"context"
	cryptoRand "crypto/rand"
	"fmt"
	"math/big"
	mathRand "math/rand"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/request"
	"github.com/dtroode/urlshorter/internal/service/dto"
	"github.com/dtroode/urlshorter/internal/service/mocks"
)

// Оригинальная реализация с math/rand
func generateStringOriginal(length int) string {
	var characters = []rune("ABCDEF0123456789")
	var sb strings.Builder

	for i := 0; i < length; i++ {
		randomIndex := mathRand.Intn(len(characters))
		randomChar := characters[randomIndex]
		sb.WriteRune(randomChar)
	}

	return sb.String()
}

// Реализация с crypto/rand (более безопасная)
func generateStringCrypto(length int) string {
	const characters = "ABCDEF0123456789"
	var sb strings.Builder
	sb.Grow(length)

	for i := 0; i < length; i++ {
		n, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			panic(err)
		}
		sb.WriteByte(characters[n.Int64()])
	}

	return sb.String()
}

// Реализация с использованием bytes.Buffer
func generateStringBuffer(length int) string {
	const characters = "ABCDEF0123456789"
	var buffer strings.Builder
	buffer.Grow(length)

	for i := 0; i < length; i++ {
		randomIndex := mathRand.Intn(len(characters))
		buffer.WriteByte(characters[randomIndex])
	}

	return buffer.String()
}

func generateStringBufferCrypto(length int) string {
	const characters = "ABCDEF0123456789"
	var buffer strings.Builder
	buffer.Grow(length)

	for i := 0; i < length; i++ {
		n, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			panic(err)
		}
		buffer.WriteByte(characters[n.Int64()])
	}

	return buffer.String()
}

// Реализация с предварительным выделением слайса
func generateStringPrealloc(length int) string {
	const characters = "ABCDEF0123456789"
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		randomIndex := mathRand.Intn(len(characters))
		result[i] = characters[randomIndex]
	}

	return string(result)
}

func generateStringPreallocCrypto(length int) string {
	const characters = "ABCDEF0123456789"
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		n, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(characters))))

		if err != nil {
			panic(err)
		}
		result[i] = characters[n.Int64()]
	}

	return string(result)
}

func BenchmarkGenerateString(b *testing.B) {
	lengths := []int{5, 10, 20}

	for _, length := range lengths {
		b.Run(fmt.Sprintf("Original_Length_%d", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				generateStringOriginal(length)
			}
		})

		b.Run(fmt.Sprintf("Crypto_Length_%d", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				generateStringCrypto(length)
			}
		})

		b.Run(fmt.Sprintf("Buffer_Length_%d", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				generateStringBuffer(length)
			}
		})

		b.Run(fmt.Sprintf("BufferCrypto_Length_%d", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				generateStringBufferCrypto(length)
			}
		})

		b.Run(fmt.Sprintf("Prealloc_Length_%d", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				generateStringPrealloc(length)
			}
		})

		b.Run(fmt.Sprintf("PreallocCrypto_Length_%d", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				generateStringPreallocCrypto(length)
			}
		})
	}
}

func BenchmarkGenerateStringAllocs(b *testing.B) {
	length := 10

	b.Run("Original", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			generateStringOriginal(length)
		}
	})

	b.Run("Crypto", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			generateStringCrypto(length)
		}
	})

	b.Run("Buffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			generateStringBuffer(length)
		}
	})

	b.Run("BufferCrypto", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			generateStringBufferCrypto(length)
		}
	})

	b.Run("Prealloc", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			generateStringPrealloc(length)
		}
	})

	b.Run("PreallocCrypto", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			generateStringPreallocCrypto(length)
		}
	})
}

func BenchmarkURL_CreateShortURLBatch(b *testing.B) {
	batchSizes := []int{10, 100, 1000, 10000}
	ctx := context.Background()
	userID := uuid.New()
	storage := mocks.NewURLStorage(b)

	svc := NewURL("http://localhost:8080", 10, 10, 10, storage)

	for _, batchSize := range batchSizes {
		urls := make([]*request.CreateShortURLBatch, 0)
		urlModels := make([]*model.URL, 0)

		for i := 0; i < batchSize; i++ {
			correlationID := uuid.New().String()
			originalURL := fmt.Sprintf("https://example.com/%s", correlationID)
			urls = append(urls, &request.CreateShortURLBatch{
				CorrelationID: correlationID,
				OriginalURL:   originalURL,
			})

			urlModels = append(urlModels, &model.URL{
				ShortKey:    correlationID,
				OriginalURL: originalURL,
				UserID:      userID,
			})
		}

		storage.On("SetURLs", mock.Anything, mock.Anything).Return(urlModels, nil)
		b.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				svc.CreateShortURLBatch(ctx, &dto.CreateShortURLBatch{
					URLs:   urls,
					UserID: userID,
				})
			}
		})
	}
}

func BenchmarkURL_DeleteURLs(b *testing.B) {
	batchSizes := []int{5, 10, 20}
	urlsCount := 30
	ctx := context.Background()
	userID := uuid.New()
	storage := mocks.NewURLStorage(b)

	svc := NewURL("http://localhost:8080", 10, 10, 10, storage)

	for _, batchSize := range batchSizes {
		shortKeys := make([]string, 0)
		urlMap := make(map[string]*model.URL)

		for i := 0; i < urlsCount; i++ {
			shortKey := uuid.New().String()
			shortKeys = append(shortKeys, shortKey)

			urlModel := &model.URL{
				ID:          uuid.New(),
				ShortKey:    shortKey,
				OriginalURL: fmt.Sprintf("https://example.com/%s", shortKey),
				UserID:      userID,
			}
			urlMap[shortKey] = urlModel
		}

		storage.On("GetURLs", mock.Anything, mock.Anything).Return(func(ctx context.Context, keys []string) ([]*model.URL, error) {
			result := make([]*model.URL, 0, len(keys))
			for _, key := range keys {
				if url, exists := urlMap[key]; exists {
					result = append(result, url)
				}
			}
			return result, nil
		}, nil)

		storage.On("DeleteURLs", mock.Anything, mock.Anything).Return(nil)

		b.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				svc.DeleteURLs(ctx, &dto.DeleteURLs{
					ShortKeys: shortKeys,
					UserID:    userID,
				})
			}
		})
	}
}
