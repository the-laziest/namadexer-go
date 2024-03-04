package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"go.uber.org/zap"

	"github.com/the-laziest/namadexer-go/internal/config"
	"github.com/the-laziest/namadexer-go/internal/repository"
	"github.com/the-laziest/namadexer-go/internal/repository/postgres"
	"github.com/the-laziest/namadexer-go/internal/server"
	"github.com/the-laziest/namadexer-go/internal/service"
	"github.com/the-laziest/namadexer-go/pkg/logger"
)

func main() {
	time.Local = time.UTC

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configFilePath := os.Getenv("CONFIG_PATH")

	cfgFile, err := os.ReadFile(configFilePath)
	if err != nil {
		logger.Fatal("Failed to read config.toml", zap.Error(err))
	}

	var cfg config.Config

	err = toml.Unmarshal(cfgFile, &cfg)
	if err != nil {
		logger.Fatal("Failed to parse config.toml", zap.Error(err))
	}

	bs, err := os.ReadFile("./checksums.json")
	if err != nil {
		logger.Fatal("Failed to read checksums file", zap.Error(err))
	}
	rawChecksums := make(map[string]string)
	err = json.Unmarshal(bs, &rawChecksums)
	if err != nil {
		logger.Fatal("Failed to decode checksums file", zap.Error(err))
	}
	checksums := prepareChecksums(rawChecksums)

	dbCfg := repository.Config{
		Host:              cfg.Database.Host,
		Port:              cfg.Database.Port,
		User:              cfg.Database.User,
		Password:          cfg.Database.Password,
		DbName:            cfg.Database.DbName,
		Schema:            cfg.ChainName,
		CreateIndex:       cfg.Database.CreateIndex,
		ConnectionTimeout: cfg.Database.ConnectionTimeout,
	}

	repo, err := postgres.NewRepository(ctx, dbCfg)
	if err != nil {
		logger.Fatal("Failed to init repository", zap.Error(err))
	}

	service, err := service.New(repo, checksums)
	if err != nil {
		logger.Fatal("Failed to init service", zap.Error(err))
	}

	serverCfg := server.Config{
		Port: cfg.Server.Port,
	}

	server := server.New(serverCfg, service)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			logger.Error("Serving failed", zap.Error(err))
			interrupt <- os.Interrupt
		}
	}()

	<-interrupt
	cancel()

	logger.Info("Gracefully closing...")

	closeCtx, closeCancel := context.WithTimeout(context.Background(), time.Second*20)
	defer closeCancel()

	if closeErr := server.Close(closeCtx); closeErr != nil {
		logger.Error("Closing server failed", zap.Error(closeErr))
	}
	if closeErr := repo.Close(); closeErr != nil {
		logger.Error("Closing repository failed", zap.Error(closeErr))
	}
}

func prepareChecksums(raw map[string]string) map[string]string {
	checksums := make(map[string]string, len(raw))
	for txType, hash := range raw {
		checksums[strings.Split(txType, ".")[0]] = strings.Split(hash, ".")[1]
	}
	return checksums
}
