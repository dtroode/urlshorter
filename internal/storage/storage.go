package storage

import (
	"context"

	internalerror "github.com/dtroode/urlshorter/internal/error"
)

type URL struct {
	urlmap map[string]string
}

func NewURL() *URL {
	return &URL{
		urlmap: make(map[string]string),
	}
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
