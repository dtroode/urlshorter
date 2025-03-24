package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	internalerror "github.com/dtroode/urlshorter/internal/error"
)

type URLMap map[string]string

func (m URLMap) MarshalJSON() ([]byte, error) {
	type jsonStruct struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	urlSlice := make([]jsonStruct, 0)

	for k, v := range m {
		urlSlice = append(urlSlice, jsonStruct{ShortURL: k, OriginalURL: v})
	}

	return json.Marshal(urlSlice)
}

// TODO: think about anonymous struct or at leat deduplication of jsonStruct
func (m URLMap) UnmarshalJSON(d []byte) error {
	type jsonStruct struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	urlSlice := make([]jsonStruct, 0)

	err := json.Unmarshal(d, &urlSlice)
	if err != nil {
		return err
	}

	for _, v := range urlSlice {
		m[v.ShortURL] = v.OriginalURL
	}

	return nil
}

type URL struct {
	urlmap  URLMap
	file    *os.File
	encoder *json.Encoder
}

func NewURL(filename string) (*URL, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open file for write: %w", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read from file: %w", err)
	}

	urlmap := URLMap{}

	if len(data) != 0 {
		if err := json.Unmarshal(data, &urlmap); err != nil {
			return nil, fmt.Errorf("failed to unmarshall urls map: %w", err)
		}
	}

	return &URL{
		urlmap:  urlmap,
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (s *URL) Flush() error {
	if err := s.encoder.Encode(s.urlmap); err != nil {
		return err
	}

	return s.file.Close()
}

func (s *URL) GetURL(_ context.Context, id string) (*string, error) {
	val, ok := s.urlmap[id]

	if !ok {
		return nil, internalerror.ErrNotFound
	}

	return &val, nil
}

func (s *URL) SetURL(_ context.Context, id, url string) error {
	s.urlmap[id] = url

	return nil
}
