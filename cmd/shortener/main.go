package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dtroode/urlshorter/config"
	grpcRouter "github.com/dtroode/urlshorter/internal/api/grpc/router"
	httpRouter "github.com/dtroode/urlshorter/internal/api/http/router"
	"github.com/dtroode/urlshorter/internal/api/middleware"
	"github.com/dtroode/urlshorter/internal/auth"
	"github.com/dtroode/urlshorter/internal/logger"
	"github.com/dtroode/urlshorter/internal/model"
	"github.com/dtroode/urlshorter/internal/server"
	"github.com/dtroode/urlshorter/internal/service"
	"github.com/dtroode/urlshorter/internal/storage"
	"github.com/dtroode/urlshorter/internal/storage/inmemory"
	"github.com/dtroode/urlshorter/internal/storage/postgres"
)

var (
	buildVersion = "N/A" // set by ldflags
	buildDate    = "N/A" // set by ldflags
	buildCommit  = "N/A" // set by ldflags
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	defer stop()

	config, err := config.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.NewLog(config.LogLevel)

	var urlStorage storage.Storage
	dsn := config.DatabaseDSN
	if dsn != "" {
		urlStorage, err = postgres.NewStorage(dsn)
		if err != nil {
			logger.Fatal("failed to create database storage", "error", err)
		}
		logger.Debug("using database storage")
	} else {
		urlStorage, err = inmemory.NewStorage(config.FileStoragePath)
		if err != nil {
			logger.Fatal("failed to create inmemory storage", "error", err, "file", config.FileStoragePath)
		}
		logger.Debug("using inmemory storage")
	}
	defer func() {
		if err = urlStorage.Close(); err != nil {
			logger.Error("failed to close storage", "error", err)
		}
	}()

	urlService := service.NewURL(config.BaseURL, config.ShortKeyLength, config.ConcurrencyLimit, config.QueueSize, urlStorage)
	healthService := service.NewHealth(urlStorage)
	defer urlService.Close()

	jwt := auth.NewJWT(config.JWTSecretKey)

	var servers []model.Server
	var wg sync.WaitGroup

	httpServer, err := registerHTTPServer(
		config.HTTPRunAddr,
		urlService,
		healthService,
		jwt,
		logger,
		config.TrustedSubnet,
	)
	if err != nil {
		logger.Fatal("failed to init http server", "error", err)
	}
	servers = append(servers, httpServer)

	grpcServer := registerGRPCServer(
		config.GRPCRunAddr,
		urlService,
		jwt,
		logger,
	)
	servers = append(servers, grpcServer)

	var sl model.SecurityLayer

	if config.EnableHTTPS {
		sl = server.NewTLSListener(config.CertFileName, config.PrivateKeyFileName)
	} else {
		sl = server.NewPlainListener()
	}

	for _, server := range servers {
		wg.Add(1)
		go func(s model.Server) {
			defer wg.Done()
			logger.Info("server started", "address", s.Address())
			err := s.Start(sl)
			if err != nil && err != http.ErrServerClosed {
				logger.Error("error running server", "error", err, "address", s.Address())
			}
		}(server)
	}

	logAppVersion()

	<-ctx.Done()
	logger.Info("received interruption signal, shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, server := range servers {
		if err := server.Stop(shutdownCtx); err != nil {
			logger.Error("error during server shutdown", "error", err, "address", server.Address())
		}
	}

	wg.Wait()
	logger.Info("shutdown completed")

}

func logAppVersion() {
	tmpl := `
Build version: %s
Build date: %s
Build commit: %s
`

	fmt.Printf(tmpl, buildVersion, buildDate, buildCommit)
}

func registerGRPCServer(
	addr string,
	urlService *service.URL,
	tokenManager middleware.Token,
	logger *logger.Logger,
) model.Server {
	r := grpcRouter.New(urlService, tokenManager, logger)
	s := r.Register()

	return server.NewGRPCServer(s, addr)
}

func registerHTTPServer(
	addr string,
	urlService *service.URL,
	healthService *service.Health,
	tokenManager middleware.Token,
	logger *logger.Logger,
	trustedSubnet string,
) (model.Server, error) {
	r := httpRouter.New(
		urlService,
		healthService,
		tokenManager,
		logger,
		trustedSubnet,
	)

	err := r.Register()
	if err != nil {
		return nil, fmt.Errorf("failed to register routes: %w", err)
	}

	return server.NewHTTPServer(
		&http.Server{
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		addr,
	), nil
}
