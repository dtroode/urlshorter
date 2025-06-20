package inmemory

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/storage"
)

// File defines interface for file operations.
type File interface {
	io.WriteCloser
	io.StringWriter
}

// URLMap represents a map of short keys to URL models.
type URLMap map[string]*model.URL

// UnmarshalJSON unmarshals JSON data into URLMap.
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

// Storage represents in-memory storage implementation.
type Storage struct {
	urlmap  URLMap
	mu      sync.RWMutex
	file    File
	encoder *json.Encoder
}

// Ping checks if the storage is available.
func (s *Storage) Ping(ctx context.Context) error {
	return nil
}

// NewStorage creates new in-memory storage instance.
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

// Close closes the storage and underlying file.
func (s *Storage) Close() error {
	return s.file.Close()
}

// GetURL retrieves a URL by its short key.
func (s *Storage) GetURL(_ context.Context, shortKey string) (*model.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.urlmap[shortKey]

	if !ok {
		return nil, storage.ErrNotFound
	}

	return val, nil
}

// GetURLs retrieves multiple URLs by their short keys.
func (s *Storage) GetURLs(_ context.Context, shortKeys []string) ([]*model.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urls := make([]*model.URL, 0)

	for _, shortKey := range shortKeys {
		for _, u := range s.urlmap {
			if u.ShortKey == shortKey {
				urls = append(urls, u)
			}
		}
	}

	return urls, nil
}

// GetURLsByUserID retrieves all URLs created by a specific user.
func (s *Storage) GetURLsByUserID(ctx context.Context, userID uuid.UUID) ([]*model.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urls := make([]*model.URL, 0)

	for _, url := range s.urlmap {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

// saveToFile saves a URL to the underlying file.
func (s *Storage) saveToFile(_ context.Context, url *model.URL) error {
	return s.encoder.Encode(url)
}

// saveToFileBatch saves multiple URLs to the underlying file.
func (s *Storage) saveToFileBatch(_ context.Context, urls string) error {
	_, err := s.file.WriteString(urls)
	return err
}

// SetURL stores a single URL in the storage.
func (s *Storage) SetURL(ctx context.Context, url *model.URL) (*model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.urlmap[url.ShortKey] = url

	if err := s.saveToFile(ctx, url); err != nil {
		return nil, fmt.Errorf("failed to encode url to file: %w", err)
	}

	return url, nil
}

// SetURLs stores multiple URLs in the storage.
func (s *Storage) SetURLs(ctx context.Context, urls []*model.URL) ([]*model.URL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var builder strings.Builder

	for _, url := range urls {
		s.urlmap[url.ShortKey] = url

		b, err := json.Marshal(url)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal url: %w", err)
		}

		if _, err := builder.WriteString(string(b) + "\n"); err != nil {
			return nil, fmt.Errorf("failed write url to buffer: %w", err)
		}
	}

	if err := s.saveToFileBatch(ctx, builder.String()); err != nil {
		return nil, fmt.Errorf("failed to encode urls to file: %w", err)
	}

	return urls, nil
}

// DeleteURLs marks the specified URLs as deleted.
func (s *Storage) DeleteURLs(_ context.Context, _ []uuid.UUID) error {
	return nil
}
