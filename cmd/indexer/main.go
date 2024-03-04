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
	"github.com/the-laziest/namadexer-go/internal/indexer"
	"github.com/the-laziest/namadexer-go/internal/repository"
	"github.com/the-laziest/namadexer-go/internal/repository/postgres"
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

	if err = repo.CreateTables(ctx); err != nil {
		logger.Fatal("Failed to create database tables", zap.Error(err))
	}

	indexerCfg := indexer.Config{
		RpcURL:             cfg.Indexer.RPC,
		Checksums:          make(map[string]string),
		WaitForBlock:       cfg.Indexer.WaitForBlock,
		MaxBlocksInChannel: cfg.Indexer.MaxBlocksInChannel,
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
	indexerCfg.Checksums = prepareChecksums(rawChecksums)

	indexer, err := indexer.New(indexerCfg, repo)
	if err != nil {
		logger.Fatal("Indexer init failed", zap.Error(err))
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Indexer starting...")

	go func() {
		err := indexer.Start(ctx)
		if err != nil {
			logger.Error("Indexer run failed", zap.Error(err))
			interrupt <- os.Interrupt
		}
	}()

	<-interrupt
	cancel()

	logger.Info("Gracefully closing...")

	indexer.Close()
	if closeErr := repo.Close(); closeErr != nil {
		logger.Error("Closing repository failed", zap.Error(closeErr))
	}
}

func prepareChecksums(raw map[string]string) map[string]string {
	checksums := make(map[string]string, len(raw))
	for txType, hash := range raw {
		checksums[strings.Split(hash, ".")[1]] = strings.Split(txType, ".")[0]
	}
	return checksums
}
