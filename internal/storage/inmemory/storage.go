package inmemory

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/storage"
)

type URLMap map[string]*model.URL

func (m URLMap) UnmarshalJSON(d []byte) error {
	urlSlice := make([]model.URL, 0)

	err := json.Unmarshal(d, &urlSlice)
	if err != nil {
		return err
	}

	for _, v := range urlSlice {
		m[v.ShortKey] = &v
	}

	return nil
}

type Storage struct {
	urlmap  URLMap
	mu      sync.RWMutex
	file    io.WriteCloser
	encoder *json.Encoder
}

func NewStorage(filename string) (*Storage, error) {
	readFile, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for read: %w", err)
	}
	defer readFile.Close()

	scanner := bufio.NewScanner(readFile)

	urlmap := URLMap{}

	for scanner.Scan() {
		entry := &model.URL{}
		if err := json.Unmarshal(scanner.Bytes(), entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshall urls entry: %w", err)
		}
		urlmap[entry.ShortKey] = entry
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	writeFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for append: %w", err)
	}

	return &Storage{
		urlmap:  urlmap,
		file:    writeFile,
		encoder: json.NewEncoder(writeFile),
	}, nil
}

func (s *Storage) Close() error {
	return s.file.Close()
}

func (s *Storage) GetURL(_ context.Context, shortKey string) (*model.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.urlmap[shortKey]

	if !ok {
		return nil, storage.ErrNotFound
	}

	return val, nil
}

func (s *Storage) GetURLByOriginal(ctx context.Context, originalURL string) (*model.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, url := range s.urlmap {
		if url.OriginalURL == originalURL {
			return url, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (s *Storage) saveToFile(_ context.Context, url *model.URL) error {
	return s.encoder.Encode(url)
}

func (s *Storage) SetURL(ctx context.Context, url *model.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urlmap[url.ShortKey] = url

	if err := s.saveToFile(ctx, url); err != nil {
		return fmt.Errorf("failed to encode url to file: %w", err)
	}

	return nil
}

func (s *Storage) SetURLs(ctx context.Context, urls []*model.URL) (savedURLs []*model.URL, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, url := range urls {
		s.urlmap[url.ShortKey] = url

		if err = s.saveToFile(ctx, url); err != nil {
			return nil, fmt.Errorf("failed to encode url to file: %w", err)
		}

		savedURLs = append(savedURLs, url)
	}

	return
}
