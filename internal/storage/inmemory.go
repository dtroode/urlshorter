package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	internalerror "github.com/dtroode/urlshorter/internal/error"
)

type URLData struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLMap map[string]*URLData

func (m URLMap) UnmarshalJSON(d []byte) error {
	urlSlice := make([]URLData, 0)

	err := json.Unmarshal(d, &urlSlice)
	if err != nil {
		return err
	}

	for _, v := range urlSlice {
		m[v.ShortURL] = &v
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
		entry := &URLData{}
		if err := json.Unmarshal(scanner.Bytes(), entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshall urls entry: %w", err)
		}
		urlmap[entry.ShortURL] = entry
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

func (s *InMemory) GetURL(_ context.Context, id string) (*string, error) {
	val, ok := s.urlmap[id]

	if !ok {
		return nil, internalerror.ErrNotFound
	}

	return &val.OriginalURL, nil
}

func (s *InMemory) SetURL(ctx context.Context, id, url string) error {
	urlData := &URLData{
		ShortURL:    id,
		OriginalURL: url,
	}
	s.urlmap[id] = urlData

	if err := s.saveToFile(ctx, urlData); err != nil {
		return fmt.Errorf("failed to encode url to file: %w", err)
	}

	return nil
}

func (s *InMemory) saveToFile(_ context.Context, urlData *URLData) error {
	return s.encoder.Encode(urlData)
}
