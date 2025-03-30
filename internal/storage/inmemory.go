package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	internalerror "github.com/dtroode/urlshorter/internal/error"
	"github.com/dtroode/urlshorter/internal/model"
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

type InMemory struct {
	urlmap  URLMap
	file    io.WriteCloser
	encoder *json.Encoder
}

func NewInMemory(filename string) (*InMemory, error) {
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

	return &InMemory{
		urlmap:  urlmap,
		file:    writeFile,
		encoder: json.NewEncoder(writeFile),
	}, nil
}

func (s *InMemory) Close() error {
	return s.file.Close()
}

func (s *InMemory) GetURL(_ context.Context, shortKey string) (*string, error) {
	val, ok := s.urlmap[shortKey]

	if !ok {
		return nil, internalerror.ErrNotFound
	}

	return &val.OriginalURL, nil
}

func (s *InMemory) SetURL(ctx context.Context, url *model.URL) error {
	s.urlmap[url.ShortKey] = url

	if err := s.saveToFile(ctx, url); err != nil {
		return fmt.Errorf("failed to encode url to file: %w", err)
	}

	return nil
}

func (s *InMemory) saveToFile(_ context.Context, url *model.URL) error {
	return s.encoder.Encode(url)
}
