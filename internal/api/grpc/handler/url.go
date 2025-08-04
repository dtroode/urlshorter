package handler

import (
	"context"
	"errors"

	c "github.com/dtroode/urlshorter/internal/api/grpc/context"
	pb "github.com/dtroode/urlshorter/internal/api/grpc/proto"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/model"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// URLService defines the interface for URL shortening operations.
// It provides methods for creating, retrieving, and managing shortened URLs.
type URLService interface {
	// GetOriginalURL retrieves the original URL associated with a short key.
	// Returns the original URL string or an error if not found or deleted.
	GetOriginalURL(ctx context.Context, shortKey string) (string, error)

	// GetUserURLs retrieves all URLs created by a specific user.
	// Returns a slice of user URLs or an error if the operation fails.
	GetUserURLs(ctx context.Context, userID uuid.UUID) ([]model.GetUserURL, error)

	// CreateShortURL creates a new shortened URL.
	// Returns the shortened URL string or an error if creation fails.
	CreateShortURL(ctx context.Context, params model.CreateShortURLParams) (string, error)

	// CreateShortURLBatch creates multiple shortened URLs in a single operation.
	// Returns a slice of created URLs with their correlation IDs or an error if creation fails.
	CreateShortURLBatch(ctx context.Context, params model.CreateShortURLBatchParams) ([]model.CreateShortURLBatch, error)

	// DeleteURLs marks the specified URLs as deleted for the given user.
	// Returns an error if the deletion operation fails.
	DeleteURLs(ctx context.Context, params model.DeleteURLsParams) error
}

// URL represents a gRPC handler for URL shortening operations.
// It implements the URLServiceServer interface and provides gRPC endpoints.
type URL struct {
	pb.UnimplementedURLServiceServer
	service        URLService
	contextManager *c.Manager
	logger         *logger.Logger
}

// NewURL creates a new URL handler instance with the provided service and logger.
// It initializes a gRPC URL handler with the URL service and context manager.
//
// Parameters:
//   - s: The URL service for business logic operations
//   - l: The logger for request logging
//
// Returns a pointer to the newly created URL handler instance.
func NewURL(s URLService, l *logger.Logger) *URL {
	return &URL{
		service:        s,
		contextManager: c.NewManager(),
		logger:         l,
	}
}

// GetOriginalURL handles gRPC requests to retrieve the original URL for a given short key.
// It calls the URL service to get the original URL and returns appropriate gRPC status codes.
//
// Parameters:
//   - ctx: The gRPC context
//   - in: The request containing the short key
//
// Returns the response with the original URL or an error with appropriate gRPC status.
func (h URL) GetOriginalURL(ctx context.Context, in *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	var response pb.GetOriginalURLResponse

	originalURL, err := h.service.GetOriginalURL(ctx, in.Id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "failed to found url with this key")
		}
		if errors.Is(err, model.ErrGone) {
			return nil, status.Error(codes.NotFound, "url is gone")
		}
		h.logger.Error("service error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	response.OriginalUrl = originalURL

	return &response, nil
}

// CreateShortURL handles gRPC requests to create a new shortened URL.
// It extracts the user ID from context and calls the URL service to create the short URL.
//
// Parameters:
//   - ctx: The gRPC context
//   - in: The request containing the original URL
//
// Returns the response with the shortened URL or an error with appropriate gRPC status.
func (h URL) CreateShortURL(ctx context.Context, in *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	var response pb.CreateShortURLResponse

	md, _ := metadata.FromIncomingContext(ctx)
	h.logger.Info("md in handler", "md", md)

	userID, ok := h.contextManager.GetUserIDFromContext(ctx)
	if !ok {
		h.logger.Error("user id not found in context")
		return nil, status.Error(codes.Internal, "internal error")
	}

	shortURL, err := h.service.CreateShortURL(ctx, model.CreateShortURLParams{
		UserID:      userID,
		OriginalURL: in.OriginalUrl,
	})
	if errors.Is(err, model.ErrConflict) {
		response.ShortUrl = shortURL
		response.Existed = true
		return &response, nil
	}
	if err != nil {
		h.logger.Error("service error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	response.ShortUrl = shortURL
	response.Existed = false

	return &response, nil
}

// GetUserURLs handles gRPC requests to retrieve all URLs created by a specific user.
// It extracts the user ID from context and calls the URL service to get user URLs.
//
// Parameters:
//   - ctx: The gRPC context
//   - _: The request (unused)
//
// Returns the response with user URLs or an error with appropriate gRPC status.
func (h URL) GetUserURLs(ctx context.Context, _ *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	var response pb.GetUserURLsResponse

	userID, ok := h.contextManager.GetUserIDFromContext(ctx)
	if !ok {
		h.logger.Error("user id not found in context")
		return nil, status.Error(codes.Internal, "internal error")
	}

	userURLs, err := h.service.GetUserURLs(ctx, userID)
	if err != nil {
		if errors.Is(err, model.ErrNoContent) {
			return nil, status.Error(codes.NotFound, "no urls")
		}

		h.logger.Error("service error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	response.Urls = make([]*pb.URL, 0, len(userURLs))
	for _, url := range userURLs {
		response.Urls = append(response.Urls, &pb.URL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return &response, nil
}

// DeleteURLs handles gRPC requests to delete URLs for a specific user.
// It extracts the user ID from context and calls the URL service to delete the specified URLs.
//
// Parameters:
//   - ctx: The gRPC context
//   - in: The request containing the short URLs to delete
//
// Returns the response or an error with appropriate gRPC status.
func (h URL) DeleteURLs(ctx context.Context, in *pb.DeleteURLsRequest) (*pb.DeleteURLsResponse, error) {
	var response pb.DeleteURLsResponse

	userID, ok := h.contextManager.GetUserIDFromContext(ctx)
	if !ok {
		h.logger.Error("user id not found in context")
		return nil, status.Error(codes.Internal, "internal error")
	}

	err := h.service.DeleteURLs(ctx, model.DeleteURLsParams{
		UserID:    userID,
		ShortKeys: in.ShortUrls,
	})
	if err != nil {
		h.logger.Error("failed to delete urls", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &response, nil
}
